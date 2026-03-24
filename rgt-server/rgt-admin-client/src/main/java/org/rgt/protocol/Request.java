/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol;

import org.rgt.Operation;

/**
 *
 * @author fabio_uggeri
 */
public class Request {

   private final Operation operation;

   public Request(Operation operation) {
      this.operation = operation;
   }

   public Operation getOperation() {
      return operation;
   }
}
