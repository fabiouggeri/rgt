/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.login;

import org.rgt.ServerStatus;
import org.rgt.admin.AdminOperation;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.Response;

/**
 *
 * @author fabio_uggeri
 */
public class LoginResponse extends Response {

   private ServerStatus serverStatus;

   private int sessionsCount;

   private long startTime;

   private boolean readOnly;

   private short adminProtocolVersion = AdminOperation.ADMIN_PROTOCOL_VERSION;

   private String serverVersion = "";

   private String userEditing = "";

   public LoginResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
      this.readOnly = false;
   }

   public LoginResponse(AdminResponseCode responseCode) {
      this(responseCode, null);
   }

   public LoginResponse() {
      this(AdminResponseCode.SUCCESS, null);
   }

   public LoginResponse(ServerStatus serverStatus, int sessionsCount, long startTime) {
      this();
      this.serverStatus = serverStatus;
      this.sessionsCount = sessionsCount;
      this.startTime = startTime;
   }

   /**
    * @return the serverStatus
    */
   public ServerStatus serverStatus() {
      return serverStatus;
   }

   /**
    * @param serverStatus the serverStatus to set
    * @return the object itself
    */
   public LoginResponse serverStatus(ServerStatus serverStatus) {
      this.serverStatus = serverStatus;
      return this;
   }

   /**
    * @return the sessionsCount
    */
   public int sessionsCount() {
      return sessionsCount;
   }

   /**
    * @param sessionsCount the sessionsCount to set
    * @return the object itself
    */
   public LoginResponse sessionsCount(int sessionsCount) {
      this.sessionsCount = sessionsCount;
      return this;
   }

   /**
    * @return the startTime
    */
   public long startTime() {
      return startTime;
   }

   /**
    * @param startTime the startTime to set
    * @return the object itself
    */
   public LoginResponse startTime(long startTime) {
      this.startTime = startTime;
      return this;
   }

   /**
    * Is this admin session read only?
    * @return the object itself
    */
   public boolean readOnly() {
      return readOnly;
   }

   /**
    * Set this admin session as read only or no
    * @param readOnly
    * @return the object itself
    */
   public LoginResponse readOnly(boolean readOnly) {
      this.readOnly = readOnly;
      return this;
   }

   public short adminProtocolVersion() {
      return adminProtocolVersion;
   }

   public LoginResponse adminProtocolVersion(short version) {
      this.adminProtocolVersion = version;
      return this;
   }

   public String serverVersion() {
      return serverVersion;
   }

   public LoginResponse serverVersion(String serverVersion) {
      this.serverVersion = serverVersion;
      return this;
   }

   public String userEditing() {
      return userEditing;
   }

   public LoginResponse userEditing(String userEditing) {
      this.userEditing = userEditing;
      return this;
   }

}
