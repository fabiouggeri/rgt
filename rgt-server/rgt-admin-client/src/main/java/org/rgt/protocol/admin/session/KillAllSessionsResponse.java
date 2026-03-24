/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.session;

import org.rgt.protocol.Response;
import org.rgt.admin.AdminResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public class KillAllSessionsResponse extends Response {

   private int killedSessions;

   public KillAllSessionsResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public KillAllSessionsResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public KillAllSessionsResponse(final int killedSessions) {
      super(AdminResponseCode.SUCCESS);
      this.killedSessions = killedSessions;
   }

   public KillAllSessionsResponse() {
      this(0);
   }

   public int killedSessions() {
      return killedSessions;
   }

   public KillAllSessionsResponse killedSessions(int killedSessions) {
      this.killedSessions = killedSessions;
      return this;
   }

}
