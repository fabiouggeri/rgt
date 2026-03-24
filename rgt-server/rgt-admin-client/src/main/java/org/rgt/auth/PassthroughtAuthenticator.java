package org.rgt.auth;


public class PassthroughtAuthenticator implements UserAuthenticator {

   @Override
   public boolean authenticate(final String username, final String password) throws AuthenticatorException {
      return true;
   }

   @Override
   public String encrypt(String str) throws AuthenticatorException {
      return str;
   }

   @Override
   public String decrypt(String str) throws AuthenticatorException {
      return str;
   }

}