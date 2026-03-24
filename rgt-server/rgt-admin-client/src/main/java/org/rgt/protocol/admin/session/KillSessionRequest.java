/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.session;

import org.rgt.protocol.Request;
import org.rgt.admin.AdminOperation;

/**
 *
 * @author fabio_uggeri
 */
public class KillSessionRequest extends Request {

   private long sessionId;

   public KillSessionRequest(long sessionId) {
      super(AdminOperation.KILL_SESSION);
      this.sessionId = sessionId;
   }

   public KillSessionRequest() {
      this(0L);
   }

   public long sessionId() {
      return sessionId;
   }

   public KillSessionRequest sessionId(long sessionId) {
      this.sessionId = sessionId;
      return this;
   }
}
