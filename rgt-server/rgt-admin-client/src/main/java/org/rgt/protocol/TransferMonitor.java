/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol;

/**
 *
 * @author fabio_uggeri
 */
@FunctionalInterface
public interface TransferMonitor {
   boolean transfer(int size);
}
