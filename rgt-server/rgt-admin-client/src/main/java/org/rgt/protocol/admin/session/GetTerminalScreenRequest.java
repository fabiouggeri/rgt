/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.session;

import org.rgt.admin.AdminOperation;
import org.rgt.protocol.Request;

/**
 *
 * @author fabio_uggeri
 */
public class GetTerminalScreenRequest extends Request {

   private long sessionId;

   public GetTerminalScreenRequest(final long sessionId) {
      super(AdminOperation.SEND_TERMINAL_REQUEST);
      this.sessionId = sessionId;
   }

   public GetTerminalScreenRequest() {
      this(0L);
   }

   public long sessionId() {
      return sessionId;
   }

   public GetTerminalScreenRequest sessionId(long sessionId) {
      this.sessionId = sessionId;
      return this;
   }

}
