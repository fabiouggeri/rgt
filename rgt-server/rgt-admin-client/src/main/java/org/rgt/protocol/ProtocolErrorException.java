/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol;

import org.rgt.TerminalException;

/**
 *
 * @author fabio_uggeri
 */
public class ProtocolErrorException extends TerminalException {

   private static final long serialVersionUID = -4883402625177108144L;

   public ProtocolErrorException(final String message) {
      super(null, message != null ? message : "Protocol error");
   }

   public ProtocolErrorException() {
      this(null);
   }
}
