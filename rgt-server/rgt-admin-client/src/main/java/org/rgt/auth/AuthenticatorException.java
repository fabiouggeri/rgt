/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.auth;

import org.rgt.TerminalException;
import org.rgt.admin.AdminResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public class AuthenticatorException extends TerminalException {
   
   public AuthenticatorException(AdminResponseCode error, String message, Exception ex) {
      super(error, message, ex);
   }

   public AuthenticatorException(AdminResponseCode error, Exception ex) {
      this(error, null, ex);
   }

   public AuthenticatorException(AdminResponseCode error, String message) {
      this(error, message, null);
   }

   public AuthenticatorException(AdminResponseCode error) {
      this(error, null, null);
   }
   
}
