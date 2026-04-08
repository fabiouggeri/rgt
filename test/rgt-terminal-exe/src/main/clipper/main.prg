#Include "fileio.ch"
#Include "setcurs.ch"

#IfDef __HBR__
   #Include "hbgtinfo.ch"
#EndIf

REQUEST TONE, FRENAME

// Request F_SenhaTFHSM
Request NetName
Request SIMULA_INTERACAO
Request INTERROMPE_SIMULACAO_INTERACAO
Request FERASE


#IfDef __HBR__
   Request HB_UserName
#EndIf

#Define MATCH_OPTION( a, o ) Left( a, Len( o ) ) == o
#Define GET_OPTION( o, a ) SubStr( a, Len( o ) + 1 )


procedure main(...)
   Local aParams := hb_AParams()
   Local cHost
   Local cPort
   Local cWorkDir
   Local cVersao
   Local aArgs := {}
   Local nArg
   Local cArg
   Local cOption
   Local nWDPos
   Local cUser := "guest"
   Local cPswd := ""
   Local cCmdLine

   #IfDef __HBR__
      REQUEST HB_CODEPAGE_PT850
      HB_CdpSelect("PT850")
      HB_SetTermCP("PT850")
      HB_GtInfo( HB_GTI_MOUSESTATUS, .F. )
   #EndIf

   cVersao   := "1.0.0"
   cHost     := IIf( ! Empty( GetEnv( "RGT_SERVER" ) )     , AllTrim( GetEnv( "RGT_SERVER" ) )     , "127.0.0.1" )
   cPort     := IIf( ! Empty( GetEnv( "RGT_SERVER_PORT" ) ), AllTrim( GetEnv( "RGT_SERVER_PORT" ) ), "7654" )

   SET( _SET_EXACT      , .T. )
   SET( _SET_DELETED    , .T. )
   SET( _SET_WRAP       , .T. )
   SET( _SET_BELL       , .T. )
   SET( _SET_CANCEL     , .F. )
   SET( _SET_ESCAPE     , .T. )
   SET( _SET_SCOREBOARD , .F. )
   SET( _SET_DATEFORMAT, "DD/MM/YYYY" )
   SET( _SET_EPOCH, 1920 )

   ?
   ? "Remote Graphical Terminal"
   ?
   ? " *** Terminal Client - version " + cVersao
   ? "     Host ..: " + cHost
   
   //Rgt_LogLevel(5)
   
   If PCount() < 2

      ? "Syntax:"
      ? "         terminal [options] workingDirectory programName [parameters]"
      ?
      ?
      Quit

   EndIf

   For nArg := 1 to Len( aParams )

      cArg := aParams[ nArg ]

      If Left( cArg, 1 ) == "-"
  
         cOption := Upper( cArg )

         If MATCH_OPTION( cOption, "-SERVER=" )

            cHost := GET_OPTION( "-SERVER=", cArg )

         ElseIf MATCH_OPTION( cOption, "-PORT=" )

            cPort := GET_OPTION( "-PORT=", cArg )

         ElseIf MATCH_OPTION( cOption, "-USER=" )

            cUser := GET_OPTION( "-USER=", cArg )

         ElseIf MATCH_OPTION( cOption, "-PASSWD=" )

            cPswd := GET_OPTION( "-PASSWD=", cArg )

         EndIf

         ? "Invalid option: " + cArg
         Quit

      Else

         nWDPos := nArg
         Exit

      EndIf

   Next

   cWorkDir    := aParams[ nWDPos     ]
   cCmdLine    := aParams[ nWDPos + 1 ]

   For nArg := nWDPos + 2 to Len( aParams )
      AAdd( aArgs, aParams[ nArg ] )
   Next

   #ifndef __LNX__
      cWorkDir := StrTran( cWorkDir   , "/", "\" ) // Linux fix
      cCmdLine := StrTran( cCmdLine, "/", "\" ) // Linux fix
   #endif

   If Empty( cHost )
      ? "RGT Server address not provided."
      Quit
   EndIf

   rgt_TrmDefaultUpdateInterval(350)
   rgt_TrmMaxKeysBuffer(3)


   rgt_TrmLogin( cHost, Val( cPort ), cCmdLine, cWorkDir, cUser, cPswd, aArgs )

   // process application updates and user keystrokes
   rgt_TrmListen()

   SetCursor( SC_NORMAL )

   // disconnect from application and make cleanup
   rgt_TrmLogout()

Return

Function HB_NOMOUSE()
Return( NIL )

#IfDef __XHB__
Function HB_UserName()
Return NetName(1)
#EndIf

Function Teste()
   QOut("teste")
Return Nil

Function FuncaoRemota( cMensagem )
   Alert( "Mensagem: " + cMensagem )
Return "Sucesso"

Function __DBGENTRY()
Return Nil


#PRAGMA BEGINDUMP

#ifdef __linux__

   #include <stdio.h>
   #include <unistd.h>
   #include <errno.h>
   #include <regex.h>
   #include <dirent.h>
   #include <fcntl.h>
   #include <linux/input.h>
   #include <stdbool.h>
   #include <stdint.h>
   #include <sys/time.h>

#else
   #include "windows.h"
#endif

   #include "cfl_thread.h"
   #include "hbapi.h"

   static int s_loopTeclas = 1;
   static char *s_buffer;
   static int s_bufferLen = 0;
   static int s_freeBuffer = 0;
   static unsigned int s_intervalo;

#ifdef __linux__
   static void simulateKeyPress(int conIn, int key) {
     struct input_event forcedKey;
     forcedKey.type = EV_KEY;
     forcedKey.value = 1;    // Press
     forcedKey.code = key;
     gettimeofday(&forcedKey.time,NULL);
     write(conIn, &forcedKey, sizeof(struct input_event));
   }
#else
   static void simulateKeyPress(HANDLE conIn, int key) {
      DWORD dw;
      INPUT_RECORD record[2];
      record[0].EventType = KEY_EVENT;
      record[0].Event.KeyEvent.bKeyDown = TRUE;
      record[0].Event.KeyEvent.dwControlKeyState = 0;
      record[0].Event.KeyEvent.uChar.AsciiChar = key;
      record[0].Event.KeyEvent.wRepeatCount = 1;
      record[0].Event.KeyEvent.wVirtualKeyCode = toupper(key); /* virtual keycode is always uppercase */
      record[0].Event.KeyEvent.wVirtualScanCode = MapVirtualKeyA(key & 0x00ff, 0);
      WriteConsoleInput(conIn, record, 1 , &dw);
   }
#endif

   static void loopTeclas(void *param) {
#ifdef __linux__
      int conIn = STDIN_FILENO;
//      int conIn = -1;
//      char *dirName = "/dev/input/by-id";
//      DIR *dirp;
//      struct dirent *dp;
//      char fullPath[1024];
//      int result;
//      regex_t kbd;
//
//      if(regcomp(&kbd,"event-kbd",0)!=0) {
//         printf("regcomp for kbd failed\n");
//         return 0;
//      }
//
//      if ((dirp = opendir(dirName)) == NULL) {
//         perror("couldn't open '/dev/input/by-path'");
//         return 0;
//      }
//
//
//      do {
//         errno = 0;
//         if ((dp = readdir(dirp)) != NULL) {
//            if(regexec (&kbd, dp->d_name, 0, NULL, 0) == 0) {
//               sprintf(fullPath,"%s/%s",dirName,dp->d_name);
//               conIn = open(fullPath,O_WRONLY | O_NONBLOCK);
//               result = ioctl(conIn, EVIOCGRAB, 1);
//            }
//
//         }
//      } while (dp != NULL);
//
//      closedir(dirp);
//
//
//      regfree(&kbd);

#else
      HANDLE conIn = GetStdHandle(STD_INPUT_HANDLE);
#endif
      int i = 0;
      while (s_loopTeclas) {
         simulateKeyPress(conIn, (int) s_buffer[ i % s_bufferLen ]);
         ++i;
         #ifdef __linux__
            sleep( s_intervalo );
         #else
            Sleep( s_intervalo );
         #endif
      }
      if (s_freeBuffer) {
         hb_xfree(s_buffer);
      }
#ifdef __linux__
      close(conIn);
#endif
   }

   HB_FUNC( SIMULA_INTERACAO )
   {
      CFL_THREADP hThread;
      PHB_ITEM pBuffer = hb_param( 1, HB_IT_STRING );
      PHB_ITEM pIntervalo = hb_param( 2, HB_IT_NUMERIC );

      if (pBuffer == NULL || hb_itemGetCLen(pBuffer) == 0) {
         s_buffer = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
         s_bufferLen = 26;
         s_freeBuffer = 0;
      } else {
         s_bufferLen = (int) hb_itemGetCLen(pBuffer);
         s_buffer = ( char * ) hb_xgrab( s_bufferLen + 1 );
         strncpy(s_buffer, hb_itemGetCPtr(pBuffer), s_bufferLen);
         s_buffer[ s_bufferLen ] = '\0';
         s_freeBuffer = 1;
      }

      if (pIntervalo == NULL || hb_itemGetNI(pIntervalo) == 0) {
         s_intervalo = 100;
      } else {
         s_intervalo = hb_itemGetNI(pIntervalo);
      }

      s_loopTeclas = 1;
      hThread = cfl_thread_new(loopTeclas);
      cfl_thread_start(hThread, NULL);
   }

   HB_FUNC( INTERROMPE_SIMULACAO_INTERACAO )
   {
      s_loopTeclas = 0;
   }

#PRAGMA ENDDUMP
