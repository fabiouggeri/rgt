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
public enum ServerStatus {
   STOPPED {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Server.status.stopped");
      }

   },
   STARTING {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Server.status.starting");
      }

   },
   RUNNING {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Server.status.running");
      }

   },
   STOPPING {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Server.status.stopping");
      }

   },
   DISCONNECTED {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Server.status.disconnected");
      }

   },
   CONNECTING {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Server.status.connecting");
      }

   },
   DISCONNECTING {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Server.status.disconnecting");
      }

   },
   UNKNOWN {
      @Override
      public String toString() {
         return TerminalUtil.getMessage("Server.status.unknown");
      }

   };
   public static ServerStatus getByName(final String statusName) {
      for (final ServerStatus s : values()) {
         if (s.name().equalsIgnoreCase(statusName)) {
            return s;
         }
      }
      return UNKNOWN;
   }   
}
