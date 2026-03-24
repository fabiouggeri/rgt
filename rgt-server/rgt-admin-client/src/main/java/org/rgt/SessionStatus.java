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
   }
}
