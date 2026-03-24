/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.users;

import java.util.ArrayList;
import java.util.Collection;
import org.rgt.auth.TerminalUser;
import org.rgt.protocol.Response;
import org.rgt.admin.AdminResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public class GetUsersResponse extends Response {

   private boolean userAuthentication = true;

   private final ArrayList<TerminalUser> users = new ArrayList<>();

   public GetUsersResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public GetUsersResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public GetUsersResponse() {
      super(AdminResponseCode.SUCCESS);
   }

   public void addUsers(Collection<TerminalUser> users) {
      this.users.addAll(users);
   }

   public ArrayList<TerminalUser> getUsers() {
      return users;
   }

   public void setUserAuthentication(boolean enabled) {
      this.userAuthentication = enabled;
   }

   public boolean hasUserAuthentication() {
      return userAuthentication;
   }
}
