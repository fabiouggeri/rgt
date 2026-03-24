/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.server;

import org.rgt.ServerStatus;
import org.rgt.protocol.Response;
import org.rgt.admin.AdminResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public class ServerInfoResponse extends Response {

   private ServerStatus status;

   private int sessionsCount;

   private long startTime;

   public ServerInfoResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public ServerInfoResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public ServerInfoResponse() {
      this(AdminResponseCode.SUCCESS, null);
   }

   public ServerInfoResponse(ServerStatus status, int sessionsCount, long startTime) {
      this();
      this.status = status;
      this.sessionsCount = sessionsCount;
      this.startTime = startTime;
   }

   /**
    * @return the status
    */
   public ServerStatus status() {
      return status;
   }

   /**
    * @param status the status to set
    * @return this
    */
   public ServerInfoResponse status(ServerStatus status) {
      this.status = status;
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
    * @return
    */
   public ServerInfoResponse sessionsCount(int sessionsCount) {
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
    * @return
    */
   public ServerInfoResponse startTime(long startTime) {
      this.startTime = startTime;
      return this;
   }
}
