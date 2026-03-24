/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.auth;

import java.io.Serializable;
import java.util.Date;
import java.util.Objects;

/**
 *
 * @author fabio_uggeri
 */
public class TerminalUser implements Serializable {

   private static final long serialVersionUID = -1735723766366082402L;

   private final String username;
   private String password;
   private Date expiration = null;

   public TerminalUser(final String username, final String password) {
      this.username = username;
      this.password = password;
   }

   /**
    * @return the username
    */
   public String getUsername() {
      return username;
   }

   /**
    * @return the password
    */
   public String getPassword() {
      return password;
   }

   /**
    * @param password the password to set
    */
   public void setPassword(String password) {
      this.password = password;
   }

   /**
    * @return the expiration
    */
   public Date getExpiration() {
      return expiration;
   }

   /**
    * @param expiration the expiration to set
    */
   public void setExpiration(Date expiration) {
      this.expiration = expiration;
   }

   @Override
   public String toString() {
      return username;
   }

   @Override
   public boolean equals(Object obj) {
      if (! (obj instanceof TerminalUser)) {
         return false;
      }
      return getUsername().equals(((TerminalUser)obj).getUsername());
   }

   @Override
   public int hashCode() {
      int hash = 7;
      hash = 97 * hash + Objects.hashCode(this.username);
      return hash;
   }
}
