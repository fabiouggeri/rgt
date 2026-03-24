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
public interface ResponseCode {

   short value();
   String getMessage();
   boolean isSuccess();

   default boolean isError() {
      return ! isSuccess();
   }
}
