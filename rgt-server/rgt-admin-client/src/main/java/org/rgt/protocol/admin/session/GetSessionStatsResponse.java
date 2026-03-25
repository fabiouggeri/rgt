/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.protocol.admin.session;

import org.rgt.admin.AdminResponseCode;
import org.rgt.client.SessionStats;
import org.rgt.protocol.Response;

/**
 *
 * @author fabio
 */
public class GetSessionStatsResponse extends Response {
   
   private SessionStats sessionStats;
   
   public GetSessionStatsResponse() {
      super(AdminResponseCode.SUCCESS);
   }

   public GetSessionStatsResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public GetSessionStatsResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public SessionStats getSessionStats() {
      return sessionStats;
   }

   public void setSessionStats(SessionStats sessionStats) {
      this.sessionStats = sessionStats;
   }
}
