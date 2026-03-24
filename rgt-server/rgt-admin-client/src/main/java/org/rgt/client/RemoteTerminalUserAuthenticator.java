/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.client;

import org.rgt.TerminalException;
import org.rgt.admin.AdminResponseCode;
import org.rgt.auth.AuthenticatorException;
import org.rgt.auth.TerminalUserAuthenticator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public class RemoteTerminalUserAuthenticator extends TerminalUserAuthenticator {
   
   private static final Logger LOG = LoggerFactory.getLogger(RemoteTerminalUserAuthenticator.class);
   
   private final RemoteTerminalServer server;

   public RemoteTerminalUserAuthenticator(RemoteTerminalServer server) throws AuthenticatorException {
      this.server = server;
   }

   @Override
   public void save() throws AuthenticatorException {
      try {
         server.setRemoteTerminalUsers();
         server.saveRemoteTerminalUsers();
      } catch (TerminalException ex) {
         LOG.error("Error saving users on server " + server, ex);
         throw new AuthenticatorException(AdminResponseCode.UNKNOWN_ERROR, "Error saving users on server " + server, ex);
      }
   }

   @Override
   public void load() throws AuthenticatorException {
      try {
         server.loadRemoteTerminalUsers();
         server.updateTerminalUsers();
      } catch (TerminalException ex) {
         LOG.error("Error loading users on server " + server, ex);
         throw new AuthenticatorException(AdminResponseCode.UNKNOWN_ERROR, "Error loading users on server " + server, ex);
      }
   }
}
