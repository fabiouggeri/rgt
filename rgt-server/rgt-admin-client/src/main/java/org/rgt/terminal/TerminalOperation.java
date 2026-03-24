/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.terminal;

import org.rgt.Operation;

/**
 *
 * @author fabio_uggeri
 */
public enum TerminalOperation implements Operation {
   UNKNOWN(0, "Unknown"),
   TRM_CMD_LOGIN(0x0A, "TE Login"),
   TRM_CMD_LOGOUT(0x0B, "TE Logout"),
   TRM_CMD_RECONNECT(0x0C, "TE Reconnect"),
   
   APP_CMD_LOGIN(0x32, "App Login"),
   APP_CMD_LOGOUT(0x33, "App Logout"),
   APP_CMD_SET_ENV(0x34, "App SetEnv"),
   APP_CMD_UPDATE(0x35, "App Update"),
   // APP_CMD_READ_KEY(0x36, "App ReadKey"), --> Deprecated
   APP_CMD_RPC(0x37, "App RPC"),
   APP_CMD_PUT_FILE(0x38, "App PutFile"),
   APP_CMD_GET_FILE(0x39, "App GetFile"),
   APP_CMD_KEY_BUF_LEN(0x3A, "App KeyBufferLen"),
   APP_CMD_RECONNECT(0x3B, "App Reconnect"),
   APP_CMD_KEEP_ALIVE(0x3C, "App KeepAlive"),
   APP_SESSION_CONFIG(0x3D, "App SessionConfig"),
   
   STANDALONE_APP_EXEC(0x64, "StandaloneApp AppExec"),
   STANDALONE_APP_SEND_OUTPUT(0x65, "StandaloneApp SendOutput"),
   STANDALONE_APP_SEND_STATUS(0x66, "StandaloneApp SendStatus"),
   
   ADMIN_GET_SCREEN(0x80, "Server GetScreen"),
   ADMIN_RESPONSE(0x7E, "Admin Response");

   public static final short TERMINAL_PROTOCOL_VERSION = 2;

   private final byte code;
   private final String description;

   private TerminalOperation(final int code, final String description) {
      this.code = (byte) code;
      this.description = description;
   }

   @Override
   public byte getCode() {
      return code;
   }

   @Override
   public short getMaxVersion() {
      return TERMINAL_PROTOCOL_VERSION;
   }

   @Override
   public String toString() {
      return "TERMINAL: " + code + " - " + description;
   }

   public static TerminalOperation getByCode(byte id) {
      for (TerminalOperation c : TerminalOperation.values()) {
         if (c.code == id) {
            return c;
         }
      }
      return UNKNOWN;
   }
}
