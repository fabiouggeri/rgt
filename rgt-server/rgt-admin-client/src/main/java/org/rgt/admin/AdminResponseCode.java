/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.admin;

import org.rgt.ResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public enum AdminResponseCode implements ResponseCode {

   SUCCESS(0, "Success"),
   SERVER_ERROR(10, "Server error"),
   INVALID_STATUS(11, "Invalid serve status"),
   SESSION_NOT_FOUND(12, "Session not found"),
   ADMIN_SESSION_ALREADY_OPEN(13, "Another administrative session is open"),
   NOT_LOGGED(14, "Not logged"),
   UNKNOWN_COMMAND(15, "Unknown command"),
   ERROR_KILLING_ADMIN_SESSION(16, "Error killing admin session"),
   INVALID_CREDENTIAL(17, "Invalid credential"),
   PROTOCOL_ERROR(18, "Protocol error"),
   SOCKET(19, "Socket error"),
   CONNECTION_LOST(20, "Connection lost"),
   NOT_ALLOWED_OPERATION(21, "Operation not allowed"),
   FILE_READING_ERROR(22, "File reading error"),
   FILE_WRITING_ERROR(23, "File writing error"),
   AUTHENTICATOR_ERROR(24, "Authenticator error"),
   UNKNOWN_ERROR(Short.MAX_VALUE, "Unknown error");

   private final short value;
   private final String message;

   private AdminResponseCode(int value, String message) {
      this.value = (short) value;
      this.message = message;
   }

   @Override
   public short value() {
      return value;
   }

   @Override
   public String getMessage() {
      return message;
   }

   @Override
   public boolean isSuccess() {
      return value == SUCCESS.value;
   }

   public static AdminResponseCode getByValue(short value) {
      for (AdminResponseCode o : values()) {
         if (o.value == value) {
            return o;
         }
      }
      return AdminResponseCode.UNKNOWN_ERROR;
   }
}
