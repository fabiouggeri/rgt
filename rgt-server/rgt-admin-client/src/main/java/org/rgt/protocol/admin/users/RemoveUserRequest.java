/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.users;

import org.rgt.protocol.Request;
import org.rgt.admin.AdminOperation;

/**
 *
 * @author fabio_uggeri
 */
public class RemoveUserRequest extends Request {

   private String username;

   public RemoveUserRequest(String username) {
      super(AdminOperation.REMOVE_USER);
      this.username = username;
   }

   public RemoveUserRequest() {
      super(AdminOperation.REMOVE_USER);
   }

   public String username() {
      return username;
   }

   public RemoveUserRequest username(String username) {
      this.username = username;
      return this;
   }
}
