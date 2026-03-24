/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.auth;

/**
 *
 * @author fabio_uggeri
 */
public interface UserAuthenticator {

   boolean authenticate(final String username, final String password) throws AuthenticatorException;
   String encrypt(final String unencryptedString) throws AuthenticatorException;
   String decrypt(final String encryptedString) throws AuthenticatorException;
   
}
