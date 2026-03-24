/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import java.util.Objects;

/**
 *
 * @author fabio_uggeri
 */
public class Credential {

   private String username;

   private String password;

   public Credential(String username, String password) {
      this.username = username;
      this.password = password;
   }

   public String getUsername() {
      return username;
   }

   public void setUsername(String username) {
      this.username = username;
   }

   public String getPassword() {
      return password;
   }

   public void setPassword(String password) {
      this.password = password;
   }

   @Override
   public boolean equals(Object obj) {
      if (! (obj instanceof Credential)) {
         return false;
      }
      return getUsername().equals(((Credential)obj).getUsername());
   }

   @Override
   public int hashCode() {
      int hash = 3;
      hash = 29 * hash + Objects.hashCode(this.username);
      return hash;
   }


}
