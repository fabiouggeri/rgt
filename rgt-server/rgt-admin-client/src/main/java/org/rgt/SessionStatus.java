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
public enum SessionStatus {
   NEW {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Session.status.new");
      }
   },
   LAUNCHING_APP {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Session.status.launchingApp");
      }
   },
   CONNECTING {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Session.status.connecting");
      }
   },
   READY {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Session.status.ready");
      }
   },
   CLOSING {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Session.status.closing");
      }
   },
   CLOSED {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Session.status.closed");
      }
   },
   UNKNOWN {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Session.status.unknown");
      }

   };

   public static SessionStatus getByName(final String statusName) {
      for (final SessionStatus s : values()) {
         if (s.name().equalsIgnoreCase(statusName)) {
            return s;
         }
      }
      return UNKNOWN;
   }
}
