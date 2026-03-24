/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.login;

import org.rgt.protocol.Request;
import org.rgt.admin.AdminOperation;

/**
 *
 * @author fabio_uggeri
 */
public class LoginRequest extends Request {

   private short protocolVersion;

   private String username;

   private String password;

   public LoginRequest(short protocolVersion, String username, String password) {
      super(AdminOperation.LOGIN);
      this.protocolVersion = protocolVersion;
      this.username = username;
      this.password = password;
   }

   public LoginRequest() {
      this((short)0, null, null);
   }

   public short protocolVersion() {
      return protocolVersion;
   }

   public LoginRequest protocolVersion(short protocolVersion) {
      this.protocolVersion = protocolVersion;
      return this;
   }

   public String username() {
      return username;
   }

   public LoginRequest username(String username) {
      this.username = username;
      return this;
   }

   public String password() {
      return password;
   }

   public LoginRequest password(String password) {
      this.password = password;
      return this;
   }

}
