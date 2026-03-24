/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

/**
 *
 * @author fabio_uggeri
 */
public interface Session {

   String getApplicationAddress();

   String getTerminalAddress();

   long getId();

   long getStartTime();

   SessionStatus getStatus();

   void setStatus(SessionStatus status);

   String getTerminalUser();

   String getOSUser();

   void close();

   long getAppPid();

   void setAppPid(long pid);

   String getCommandLine();

   void setCommandLine(String cmdLine);
}
