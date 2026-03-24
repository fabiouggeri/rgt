/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol;

import org.rgt.ResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public class Response {

   private ResponseCode responseCode;

   private String message;

   public Response(ResponseCode responseCode, String message) {
      this.responseCode = responseCode;
      this.message = message;
   }

   public Response(ResponseCode responseCode) {
      this(responseCode, null);
   }

   public ResponseCode getResponseCode() {
      return responseCode;
   }

   public void setResponseCode(ResponseCode responseCode) {
      this.responseCode = responseCode;
   }

   public String getMessage() {
      return message != null ? message : responseCode.getMessage();
   }

   public void setMessage(String message) {
      this.message = message;
   }

   public boolean isError() {
      return responseCode != null && responseCode.isError();
   }

   public boolean isSuccess() {
      return responseCode != null && responseCode.isSuccess();
   }
}
