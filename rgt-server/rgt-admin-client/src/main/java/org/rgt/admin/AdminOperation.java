/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.admin;

import org.rgt.Operation;

/**
 *
 * @author fabio_uggeri
 */
public enum AdminOperation implements Operation {
   LOGIN(1, "login"),
   LOGOFF(2, "logout"),
   STOP_SERVICE(3, "stop service"),
   START_SERVICE(4, "start service"),
   GET_SESSIONS(5, "get sessions"),
   GET_STATUS(6, "get status"),
   KILL_SESSION(7, "kill session"),
   KILL_ALL_SESSIONS(8, "kill all sessions"),
   SET_CONFIG(9, "set config"),
   GET_CONFIG(10, "get config"),
   SAVE_CONFIG(11, "save config"),
   LOAD_CONFIG(12, "load config"),
   SET_LOG_LEVEL(13, "set log level"),
   GET_USERS(14, "get users"),
   SET_USERS(15, "set users"),
   SAVE_USERS(16, "save users"),
   LOAD_USERS(17, "load users"),
   ADD_USER(18, "add user"),
   REMOVE_USER(19, "remove user"),
   KILL_ADMIN_SESSIONS(20, "kill admin sessions"),
   LIST_FILES(21, "list files"),
   GET_FILE(22, "get a file from server"),
   PUT_FILE(23, "put a file on server"),
   REMOVE_FILE(24, "remove a file from server"),
   SEND_TERMINAL_REQUEST(25, "terminal request"),
   GET_STATS(26, "server stats"),
   GET_SESSION_STATS(27, "server stats"),
   CANCEL(126, "cancel"),
   UNKNOWN(Byte.MAX_VALUE, "Unknown");

   public static final short ADMIN_PROTOCOL_VERSION = 7;

   private final byte code;
   private final String description;

   private AdminOperation(final int code, final String description) {
      this.code = (byte) code;
      this.description = description;
   }

   @Override
   public byte getCode() {
      return code;
   }

   @Override
   public short getMaxVersion() {
      return ADMIN_PROTOCOL_VERSION;
   }

   @Override
   public String toString() {
      return "ADMIN: " + code + " - " + description;
   }

   public static AdminOperation getByCode(byte id) {
      for (AdminOperation c : AdminOperation.values()) {
         if (c.code == id) {
            return c;
         }
      }
      return UNKNOWN;
   }
}
