/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

/**
 *
 * @author fabio_uggeri
 */
public class TerminalException extends Exception {

   private static final long serialVersionUID = 3829983619750427257L;

   private ResponseCode error;

   public TerminalException(ResponseCode error, String message, Exception ex) {
      super(message != null ? message : error.getMessage(), ex);
      this.error = error;
   }

   public TerminalException(ResponseCode error, Exception ex) {
      this(error, null, ex);
   }

   public TerminalException(ResponseCode error, String message) {
      this(error, message, null);
   }

   public TerminalException(ResponseCode error) {
      this(error, null, null);
   }

   public ResponseCode getError() {
      return error;
   }

   @Override
   public String getMessage() {
      if (getCause() != null) {
         return super.getMessage() + "\nCaused By: " +
                 (getCause().getMessage() != null ? getCause().getMessage() : getCause().toString());
      } else {
         return super.getMessage();
      }
   }

   @Override
   public String getLocalizedMessage() {
      if (getCause() != null) {
         return super.getLocalizedMessage() + "\nCaused By: " +
                 (getCause().getMessage() != null ? getCause().getLocalizedMessage() : getCause().toString()) ;
      } else {
         return super.getLocalizedMessage();
      }
   }

}
