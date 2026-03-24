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
public interface TerminalServerListener {

   void notification(final TerminalServer server, final String message);

   void configurationLoaded(final TerminalServer server);

   void configurationSaved(final TerminalServer server);

   void statusChanged(final TerminalServer server, final ServerStatus previousStatus);

   void sessionOpen(final Session session);

   void sessionClose(final Session session);

   void serviceStart(final TerminalServer server);

   void serviceStop(final TerminalServer server);

   void stateUpdate(final TerminalServer server);

}
