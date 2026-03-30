#include "cfl_ints.h"

#include "rgt_types.h"

#include <stdlib.h>
#include <time.h>

#ifdef __linux__

#include <linux/limits.h>
#include <sys/stat.h>

#endif

#include "hbapifs.h"
#include "hbapiitm.h"
#include "hbcomp.h"
#include "hbdefs.h"
#include "hbset.h"


#include "cfl_list.h"
#include "cfl_process.h"
#include "cfl_str.h"
#include "rgt_log.h"


#define MAX_PATH 512
#define MAX_LOG_PATH_NAME_LEN MAX_PATH
#define MAX_LOG_LABEL_LEN 16

#define LOG_BUFFER_SIZE 1024 * 16

#define VAR_NAME_MAX_LEN 255
#define ENV_VAR_BEGIN '%'
#define ENV_VAR_END '%'
#define ENV_VAR_MARKS_LEN 2
#define RGT_VAR_MARK '$'
#define RGT_VAR_BEGIN '{'
#define RGT_VAR_END '}'
#define RGT_VAR_VALUE_MAX_LEN 64
#define RGT_VAR_MARKS_LEN 3

#define STR_IS_EMPTY(s) (s[0] == '\0')

#define MAX_LOG_SIZE 10 * 1024 * 1024

static char *s_logLevelsLabels[6] = {"OFF", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"};

static char s_logPathName[MAX_LOG_PATH_NAME_LEN + 1] = {0};

static char s_logLabel[MAX_LOG_LABEL_LEN + 1] = "[APP] ";

static FILE *s_logHandle = NULL;

static CFL_BOOL s_TryOpen = CFL_TRUE;

static int s_funLevel = 0;

static CFL_BOOL s_envLogLevel = CFL_FALSE;

static CFL_BOOL s_envLogFile = CFL_FALSE;

unsigned int __rgtLoglevel = RGT_LOG_LEVEL_ERROR;

/* Public variable to store returning of clock() at TRACE function */
clock_t _rgtTraceLastClock;

static size_t envVarName(char *str, int start, char *varName, size_t varMaxSize) {
   size_t i = start + 1;
   while (str[i] != '\0') {
      if (str[i] == ENV_VAR_END) {
         size_t varLen = i - start - 1;
         if (varLen > 0 && varLen < varMaxSize) {
            strncpy(varName, str + start + 1, varLen);
            varName[varLen] = '\0';
            return varLen;
         } else {
            varName[0] = '\0';
            return 0;
         }
      }
      ++i;
   }
   return 0;
}

static size_t rgtVarName(char *str, int start, char *varName, size_t varMaxSize) {
   size_t i = start + 1;
   if (str[i] != RGT_VAR_BEGIN) {
      return 0;
   }
   ++i;
   while (str[i] != '\0') {
      if (str[i] == RGT_VAR_END) {
         size_t varLen = i - start - 2;
         if (varLen > 0 && varLen < varMaxSize) {
            strncpy(varName, str + start + 2, varLen);
            varName[varLen] = '\0';
            return varLen;
         } else {
            varName[0] = '\0';
            return 0;
         }
      }
      ++i;
   }
   return 0;
}

static size_t strLength(const char *str, size_t start) {
   size_t len = start;
   while (str[len] != '\0') {
      ++len;
   }
   return len;
}

static void strReplace(char *str, size_t start, size_t len, const char *strReplace, size_t maxLen) {
   size_t strReplaceLen;
   size_t rightStrLen;
   if (start + len > maxLen) {
      return;
   }
   strReplaceLen = strlen(strReplace);
   if (start + strReplaceLen >= maxLen) {
      strReplaceLen = maxLen - start;
      rightStrLen = 0;
   } else {
      rightStrLen = strLength(str, start) - start - len;
      if (start + strReplaceLen + rightStrLen >= maxLen) {
         rightStrLen = maxLen - start - strReplaceLen;
      }
      memmove(str + start + strReplaceLen, str + start + len, rightStrLen);
   }
   memcpy(str + start, strReplace, strReplaceLen);
   str[start + strReplaceLen + rightStrLen] = '\0';
}

static CFL_BOOL rgtVarValue(const char *varName, char *varValue) {
   if (strncmp(varName, "pid", 3) == 0) {
      sprintf(varValue, "%lld", cfl_process_getId());
      return CFL_TRUE;
   } else if (strncmp(varName, "date", 4) == 0) {
      struct tm *tm = rgt_log_localtime();
      sprintf(varValue, "%04d%02d%02d", 1900 + tm->tm_year, tm->tm_mon + 1, tm->tm_mday);
      return CFL_TRUE;
   } else if (strncmp(varName, "time", 4) == 0) {
      struct tm *tm = rgt_log_localtime();
      sprintf(varValue, "%02d%02d%02d", tm->tm_hour, tm->tm_min, tm->tm_sec);
      return CFL_TRUE;
   }
   return CFL_FALSE;
}

static void replaceVariables(char *str, int maxLen) {
   int i = 0;
   while (i < maxLen && str[i] != '\0') {
      if (str[i] == ENV_VAR_BEGIN) {
         char varName[VAR_NAME_MAX_LEN + 1];
         size_t varLen = envVarName(str, i, varName, VAR_NAME_MAX_LEN);
         if (varLen > 0) {
            const char *varValue = getenv(varName);
            if (varValue != NULL) {
               strReplace(str, i, varLen + ENV_VAR_MARKS_LEN, varValue, maxLen);
            } else {
               ++i;
            }
         } else {
            ++i;
         }
      } else if (str[i] == RGT_VAR_MARK) {
         char varName[VAR_NAME_MAX_LEN + 1];
         size_t varLen = rgtVarName(str, i, varName, VAR_NAME_MAX_LEN);
         if (varLen > 0) {
            char varValue[RGT_VAR_VALUE_MAX_LEN + 1];
            if (rgtVarValue(varName, varValue)) {
               strReplace(str, i, varLen + RGT_VAR_MARKS_LEN, varValue, maxLen);
            } else {
               ++i;
            }
         } else {
            ++i;
         }
      } else {
         ++i;
      }
   }
}

static void replaceChars(char *str, char from, char to) {
   char *ptr = str;
   while (*ptr != '\0') {
      if (*ptr == from) {
         *ptr = to;
      }
      ++ptr;
   }
}

static void formatArgs(CFL_STRP pStr, const char *format, va_list pArgs) {
   char buffer[LOG_BUFFER_SIZE];
   int iLen;

   iLen = vsnprintf(buffer, LOG_BUFFER_SIZE, format, pArgs);
   if (iLen < 0) {
      iLen = 0;
   } else if (iLen >= LOG_BUFFER_SIZE) {
      iLen = LOG_BUFFER_SIZE - 1;
   }
   cfl_str_appendLen(pStr, buffer, iLen);
}

CFL_STRP rgt_log_format(const char *format, ...) {
   if (format != NULL && strlen(format) > 0) {
      CFL_STRP pStr = cfl_str_new(128);
      va_list pArgs;

      va_start(pArgs, format);
      formatArgs(pStr, format, pArgs);
      va_end(pArgs);
      return pStr;
   } else {
      return NULL;
   }
}

static char *getPathLog(char *path) {
#if defined(_WINDOWS_)
   char *envPath = getenv("TEMP");
   if (envPath != NULL) {
      size_t len;
      strncpy(path, envPath, MAX_PATH);
      len = strlen(path);
      if (len > 0 && len < MAX_PATH && path[len - 1] != '\\') {
         path[len++] = '\\';
         path[len] = '\0';
      }
   } else {
      path[0] = '\0';
   }
#elif defined(__linux__)
   char *envPath = getenv("HOME");
   if (envPath != NULL) {
      size_t len;
      strncpy(path, envPath, MAX_PATH);
      len = strlen(path);
      if (len > 0 && len < MAX_PATH && path[len - 1] != '/') {
         path[len++] = '/';
         path[len] = '\0';
      }
      if (len + 4 < MAX_PATH) {
         path[len++] = 't';
         path[len++] = 'm';
         path[len++] = 'p';
         path[len++] = '/';
         path[len] = '\0';
         mkdir(path, S_IRWXU | S_IRGRP | S_IXGRP | S_IROTH | S_IXOTH);
      }
   } else {
      strcpy(path, "/tmp/");
   }
#else
   path[0] = '\0';
#endif
   return path;
}

char *rgt_log_getPathName(void) {
   if (strlen(s_logPathName) == 0) {
      const struct tm *tm = rgt_log_localtime();
      char path[MAX_PATH] = {0};
      sprintf(s_logPathName, "%srgt_%04d%02d.log", getPathLog(path), 1900 + tm->tm_year, tm->tm_mon + 1);
   }
   return s_logPathName;
}

void rgt_log_setPathName(const char *logPathName) {
   size_t pathLen = strlen(logPathName);
   if (pathLen > 0 && pathLen <= MAX_LOG_PATH_NAME_LEN) {
      strncpy(s_logPathName, logPathName, MAX_LOG_PATH_NAME_LEN);
      s_logPathName[MAX_LOG_PATH_NAME_LEN] = '\0';
      replaceVariables(s_logPathName, MAX_LOG_PATH_NAME_LEN);
#if defined(CFL_OS_WINDOWS)
      replaceChars(s_logPathName, '/', '\\');
#elif defined(CFL_OS_LINUX)
      replaceChars(s_logPathName, '\\', '/');
#endif
      if (s_logHandle != NULL) {
         fclose(s_logHandle);
         s_logHandle = NULL;
         s_TryOpen = CFL_TRUE;
      }
   } else {
      s_logPathName[0] = '\0';
      if (s_logHandle != NULL) {
         fclose(s_logHandle);
         s_logHandle = NULL;
         s_TryOpen = CFL_TRUE;
      }
   }
}

int rgt_log_getLevel(void) {
   return __rgtLoglevel;
}

void rgt_log_setLevel(int level) {
   if (level >= RGT_LOG_LEVEL_OFF && level <= RGT_LOG_LEVEL_TRACE) {
      __rgtLoglevel = level;
      if (level == RGT_LOG_LEVEL_OFF && s_logHandle != NULL) {
         fclose(s_logHandle);
         s_logHandle = NULL;
         s_TryOpen = CFL_TRUE;
      }
   }
}

static long fileSize(const char *logFileName) {
   FILE *logHandle = fopen(logFileName, "r");
   if (logHandle != NULL) {
      long fileLen;
      fseek(logHandle, 0L, SEEK_END);
      fileLen = ftell(logHandle);
      fclose(logHandle);
      return fileLen;
   }
   return 0L;
}

static CFL_BOOL logActive(void) {
   if (__rgtLoglevel > RGT_LOG_LEVEL_OFF) {
      if (s_logHandle == NULL && s_TryOpen) {
         const char *logFileName = rgt_log_getPathName();
         if (fileSize(logFileName) > MAX_LOG_SIZE) {
            hb_fsDelete(logFileName);
         }
         s_logHandle = fopen(logFileName, "a");
         if (s_logHandle == NULL) {
            // Try again in default directory with default name pattern
            s_logPathName[0] = '\0';
            logFileName = rgt_log_getPathName();
            if (fileSize(logFileName) > MAX_LOG_SIZE) {
               hb_fsDelete(logFileName);
            }
            s_logHandle = fopen(logFileName, "a");
         }
         s_TryOpen = CFL_FALSE;
      }
      return s_logHandle != NULL;
   }
   return CFL_FALSE;
}

void rgt_log_write(int level, CFL_STRP out) {
   if (logActive()) {
      time_t curTime;
      struct tm *tm;
      time(&curTime);
      tm = localtime(&curTime);
      fprintf(s_logHandle, "%s%04d-%02d-%02dT%02d:%02d:%02d %s   %*s%s\n", s_logLabel, 1900 + tm->tm_year, tm->tm_mon + 1,
              tm->tm_mday, tm->tm_hour, tm->tm_min, tm->tm_sec, s_logLevelsLabels[level], s_funLevel, "", cfl_str_getPtr(out));
      fflush(s_logHandle);
   }
   cfl_str_free(out);
}

void rgt_log_writeEnter(const char *funName, CFL_UINT32 line, CFL_STRP out) {
   if (logActive()) {
      time_t curTime;
      struct tm *tm;
      time(&curTime);
      tm = localtime(&curTime);
      if (out) {
         fprintf(s_logHandle, "%s%04d-%02d-%02dT%02d:%02d:%02d TRACE => %*s%s(%lu) %s\n", s_logLabel, 1900 + tm->tm_year,
                 tm->tm_mon + 1, tm->tm_mday, tm->tm_hour, tm->tm_min, tm->tm_sec, s_funLevel, "", funName, line,
                 cfl_str_getPtr(out));
         cfl_str_free(out);
      } else {
         fprintf(s_logHandle, "%s%04d-%02d-%02dT%02d:%02d:%02d TRACE => %*s%s(%lu)\n", s_logLabel, 1900 + tm->tm_year,
                 tm->tm_mon + 1, tm->tm_mday, tm->tm_hour, tm->tm_min, tm->tm_sec, s_funLevel, "", funName, line);
      }
      s_funLevel += 2;
      fflush(s_logHandle);
   }
}

void rgt_log_writeExit(const char *funName, CFL_UINT32 line, CFL_STRP out) {
   if (logActive()) {
      time_t curTime;
      struct tm *tm;
      time(&curTime);
      tm = localtime(&curTime);
      s_funLevel -= 2;
      if (out) {
         fprintf(s_logHandle, "%s%04d-%02d-%02dT%02d:%02d:%02d TRACE <= %*s%s(%lu) %s\n", s_logLabel, 1900 + tm->tm_year,
                 tm->tm_mon + 1, tm->tm_mday, tm->tm_hour, tm->tm_min, tm->tm_sec, s_funLevel, "", funName, line,
                 cfl_str_getPtr(out));
         cfl_str_free(out);
      } else {
         fprintf(s_logHandle, "%s%04d-%02d-%02dT%02d:%02d:%02d TRACE <= %*s%s(%lu)\n", s_logLabel, 1900 + tm->tm_year,
                 tm->tm_mon + 1, tm->tm_mday, tm->tm_hour, tm->tm_min, tm->tm_sec, s_funLevel, "", funName, line);
      }
      fflush(s_logHandle);
   }
}

struct tm *rgt_log_localtime(void) {
   time_t curTime;
   time(&curTime);
   return localtime(&curTime);
}

void rgt_log_param(unsigned int level, CFL_INT16 iPos, PHB_ITEM pItem) {
   HB_BOOL bFreeReq;
   HB_SIZE nLen;
   char *strValue = hb_itemString(pItem, &nLen, &bFreeReq);
   LOG_OUT_FILE(level, ("par: %d type: %s value: %s", iPos, hb_itemTypeStr(pItem), strValue));
   if (bFreeReq) {
      RGT_HB_FREE(strValue);
   }
}

void rgt_log_item(unsigned int level, char *itemName, PHB_ITEM pItem) {
   HB_BOOL bFreeReq;
   HB_SIZE nLen;
   char *strValue = hb_itemString(pItem, &nLen, &bFreeReq);
   LOG_OUT_FILE(level, ("name: %s type: %s value: %s", itemName, hb_itemTypeStr(pItem), strValue));
   if (bFreeReq) {
      RGT_HB_FREE(strValue);
   }
}

static void initLogLevel(void) {
   const char *logLevel = getenv(RGT_LOG_LEVEL_VAR);
   if (logLevel != NULL) {
      int iLevel;

      if (hb_stricmp(logLevel, "OFF") == 0) {
         iLevel = RGT_LOG_LEVEL_OFF;
         s_envLogLevel = CFL_TRUE;

      } else if (hb_stricmp(logLevel, "ERROR") == 0) {
         iLevel = RGT_LOG_LEVEL_ERROR;
         s_envLogLevel = CFL_TRUE;

      } else if (hb_stricmp(logLevel, "WARN") == 0) {
         iLevel = RGT_LOG_LEVEL_WARN;
         s_envLogLevel = CFL_TRUE;

      } else if (hb_stricmp(logLevel, "INFO") == 0) {
         iLevel = RGT_LOG_LEVEL_INFO;
         s_envLogLevel = CFL_TRUE;

      } else if (hb_stricmp(logLevel, "DEBUG") == 0) {
         iLevel = RGT_LOG_LEVEL_DEBUG;
         s_envLogLevel = CFL_TRUE;

      } else if (hb_stricmp(logLevel, "TRACE") == 0) {
         iLevel = RGT_LOG_LEVEL_TRACE;
         s_envLogLevel = CFL_TRUE;

      } else {
         iLevel = atoi(logLevel);
         if (iLevel >= RGT_LOG_LEVEL_OFF && iLevel <= RGT_LOG_LEVEL_TRACE) {
            s_envLogLevel = CFL_TRUE;
         } else {
            iLevel = RGT_LOG_LEVEL_ERROR;
            s_envLogLevel = CFL_FALSE;
         }
      }
      rgt_log_setLevel(iLevel);
   }
}

static void initLogPathName(void) {
   const char *logPathName = getenv(RGT_LOG_PATH_NAME_VAR);
   if (logPathName != NULL) {
      rgt_log_setPathName(logPathName);
      s_envLogFile = CFL_TRUE;
   }
}

CFL_BOOL rgt_log_envLevel(void) {
   return s_envLogLevel;
}

CFL_BOOL rgt_log_envFile(void) {
   return s_envLogFile;
}

void rgt_log_initEnv(void) {
   initLogLevel();
   initLogPathName();
}

void rgt_log_setLabel(const char *label) {
   if (label != NULL) {
      strncpy(s_logLabel, label, MAX_LOG_LABEL_LEN);
   } else {
      s_logLabel[0] = '\0';
   }
}

HB_FUNC(RGT_LOGLEVEL) {
   PHB_ITEM pLogLevel = hb_param(1, HB_IT_NUMERIC);
   hb_retni(rgt_log_getLevel());
   if (pLogLevel) {
      rgt_log_setLevel(hb_itemGetNI(pLogLevel));
   }
}

HB_FUNC(RGT_LOGPATHNAME) {
   PHB_ITEM pLogPathName = hb_param(1, HB_IT_STRING);
   hb_retc(rgt_log_getPathName());
   if (pLogPathName) {
      rgt_log_setPathName(hb_itemGetCPtr(pLogPathName));
   }
}
