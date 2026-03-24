/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.users;

import org.rgt.auth.TerminalUser;
import org.rgt.protocol.Request;
import org.rgt.admin.AdminOperation;

/**
 *
 * @author fabio_uggeri
 */
public class AddUserRequest extends Request {

   private TerminalUser user;

   public AddUserRequest(TerminalUser user) {
      super(AdminOperation.ADD_USER);
      this.user = user;
   }

   public AddUserRequest() {
      this(null);
   }

   public TerminalUser user() {
      return user;
   }

   public AddUserRequest user(TerminalUser user) {
      this.user = user;
      return this;
   }
}
