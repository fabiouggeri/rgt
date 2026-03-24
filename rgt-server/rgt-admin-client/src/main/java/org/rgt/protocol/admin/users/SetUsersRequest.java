/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.users;

import java.util.ArrayList;
import java.util.Collection;
import org.rgt.auth.TerminalUser;
import org.rgt.protocol.Request;
import org.rgt.admin.AdminOperation;

/**
 *
 * @author fabio_uggeri
 */
public class SetUsersRequest extends Request {

   private final ArrayList<TerminalUser> users = new ArrayList<>();

   public SetUsersRequest(Collection<TerminalUser> otherUsers) {
      super(AdminOperation.SET_USERS);
      this.users.addAll(otherUsers);
   }

   public SetUsersRequest() {
      super(AdminOperation.SET_USERS);
   }

   public ArrayList<TerminalUser> users() {
      return users;
   }

   public SetUsersRequest user(TerminalUser user) {
      users.add(user);
      return this;
   }
}
