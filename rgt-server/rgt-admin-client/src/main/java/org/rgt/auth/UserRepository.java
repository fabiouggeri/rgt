/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.auth;

import java.util.Collection;

/**
 *
 * @author fabio_uggeri
 */
public interface UserRepository {

   TerminalUser addUser(final String username, final String password) throws AuthenticatorException;

   boolean addUser(final TerminalUser user) throws AuthenticatorException;

   void clearUsers() throws AuthenticatorException;

   TerminalUser findUser(final String username);

   Collection<TerminalUser> getUsers();

   void load() throws AuthenticatorException;

   TerminalUser removeUser(final String username) throws AuthenticatorException;

   void save() throws AuthenticatorException;
   
}
