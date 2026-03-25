/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.session;

import org.rgt.admin.AdminOperation;
import org.rgt.protocol.Request;

/**
 *
 * @author fabio
 */
public class GetSessionStatsRequest extends Request {

   private long sessionId;

   public GetSessionStatsRequest(long sessionId) {
      super(AdminOperation.KILL_SESSION);
      this.sessionId = sessionId;
   }

   public GetSessionStatsRequest() {
      this(0L);
   }

   public long sessionId() {
      return sessionId;
   }

   public GetSessionStatsRequest sessionId(long sessionId) {
      this.sessionId = sessionId;
      return this;
   }
}
