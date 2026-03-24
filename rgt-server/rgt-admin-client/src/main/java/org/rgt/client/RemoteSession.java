/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.client;

import org.rgt.Session;
import org.rgt.SessionStatus;

/**
 *
 * @author fabio_uggeri
 */
public class RemoteSession implements Session {

   private long id;
   private long appPid;
   private String applicationAddress;
   private String terminalAddress;
   private String terminalUser;
   private long startTime;
   private SessionStatus status;
   private String osUser;
   private String commandLine;

   public RemoteSession() {
   }

   public RemoteSession(final Session otherSession) {
      this.id = otherSession.getId();
      this.appPid = otherSession.getAppPid();
      this.applicationAddress = otherSession.getApplicationAddress();
      this.terminalAddress = otherSession.getTerminalAddress();
      this.terminalUser = otherSession.getTerminalUser();
      this.startTime = otherSession.getStartTime();
      this.status = otherSession.getStatus();
      this.osUser = otherSession.getOSUser();
      this.commandLine = otherSession.getCommandLine();
   }

   @Override
   public long getId() {
      return id;
   }

   @Override
   public long getAppPid() {
      return appPid;
   }

   @Override
   public String getApplicationAddress() {
      return applicationAddress;
   }

   @Override
   public long getStartTime() {
      return startTime;
   }

   @Override
   public SessionStatus getStatus() {
      return status;
   }

   @Override
   public String getTerminalAddress() {
      return terminalAddress;
   }

   @Override
   public String getTerminalUser() {
      return terminalUser;
   }

   /**
    * @param id the id to set
    */
   public void setId(long id) {
      this.id = id;
   }

   /**
    * @param appPid the appPid to set
    */
   @Override
   public void setAppPid(long appPid) {
      this.appPid = appPid;
   }

   /**
    * @param applicationAddress the applicationAddress to set
    */
   public void setApplicationAddress(String applicationAddress) {
      this.applicationAddress = applicationAddress;
   }

   /**
    * @param terminalAddress the terminalAddress to set
    */
   public void setTerminalAddress(String terminalAddress) {
      this.terminalAddress = terminalAddress;
   }

   /**
    * @param terminalUser the terminalUser to set
    */
   public void setTerminalUser(String terminalUser) {
      this.terminalUser = terminalUser;
   }

   /**
    * @param startTime the startTime to set
    */
   public void setStartTime(long startTime) {
      this.startTime = startTime;
   }

   /**
    * @param status the status to set
    */
   @Override
   public void setStatus(SessionStatus status) {
      this.status = status;
   }

   @Override
   public boolean equals(Object obj) {
      if (! (obj instanceof Session)) {
         return false;
      }
      return id == ((Session)obj).getId();
   }

   @Override
   public int hashCode() {
      int hash = 3;
      hash = 79 * hash + (int) (this.id ^ (this.id >>> 32));
      return hash;
   }

   @Override
   public void close() {
   }

   @Override
   public String getOSUser() {
      return osUser;
   }

   public void setOSUser(String osUser) {
      this.osUser = osUser;
   }

   @Override
   public String getCommandLine() {
      return commandLine;
   }

   @Override
   public void setCommandLine(String cmdLine) {
      this.commandLine = cmdLine;
   }
}
