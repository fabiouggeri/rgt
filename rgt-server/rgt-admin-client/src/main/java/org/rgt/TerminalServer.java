/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import org.rgt.auth.UserAuthenticator;
import java.net.InetAddress;
import java.util.Collection;
import org.rgt.auth.UserRepository;
import org.rgt.client.SessionStats;
import org.rgt.protocol.ProtocolProvider;

/**
 *
 * @author fabio_uggeri
 */
public interface TerminalServer {
   void startTerminalEmulationService() throws TerminalException;
   void stopTerminalEmulationService() throws TerminalException;
   void startAdminService() throws TerminalException;
   void stopAdminService() throws TerminalException;
   InetAddress getInetAddress();
   boolean isRunning();
   boolean isRunningAdminService();
   Collection<? extends Session> getSessions() throws TerminalException;
   int getSessionsCount();
   SessionStats getSessionStats(long sessionId) throws TerminalException;
   boolean killSession(long id) throws TerminalException;
   Session getSession(long id) throws TerminalException;
   TextScreen getTerminalScreen(long id) throws TerminalException;
   ServerStatus getStatus();
   boolean isEmbedded();
   boolean isConnected();
   void addListener(TerminalServerListener listener);
   void removeListener(TerminalServerListener listener);
   void clearListeners();
   UserAuthenticator getUserAuthenticator();
   UserRepository getUserRepository();
   void setUserAuthenticator(UserAuthenticator authenticator);
   void setUserRepository(UserRepository repository);
   TerminalServerConfiguration getConfiguration();
   long getStartTime();
   boolean saveConfiguration() throws TerminalException;
   void loadConfiguration() throws TerminalException;
   void updateLocalState() throws TerminalException;
   void updateLocalStatus() throws TerminalException;
   void updateRemoteConfiguration() throws TerminalException;
   void updateLogLevel() throws TerminalException;
   void lostConnection(Session session);
   ProtocolProvider getProtocolProvider();
   String getVersion();
   String getId();

   default boolean isReadOnly() {
      return false;
   }
}
