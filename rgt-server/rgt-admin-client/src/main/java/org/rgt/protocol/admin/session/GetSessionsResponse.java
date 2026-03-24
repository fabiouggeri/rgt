/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.session;

import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import org.rgt.client.RemoteSession;
import org.rgt.protocol.Response;
import org.rgt.admin.AdminResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public class GetSessionsResponse extends Response {

   private final ArrayList<RemoteSession> sessions = new ArrayList<>();

   public GetSessionsResponse() {
      super(AdminResponseCode.SUCCESS);
   }

   public GetSessionsResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public GetSessionsResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public void addSessions(Collection<RemoteSession> sessions) {
      this.sessions.addAll(sessions);
   }

   public void addSession(RemoteSession session) {
      sessions.add(session);
   }

   public List<RemoteSession> getSessions() {
      return sessions;
   }

}
