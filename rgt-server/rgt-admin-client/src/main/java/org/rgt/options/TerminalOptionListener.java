/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.options;

/**
 *
 * @author fabio_uggeri
 */
@FunctionalInterface
public interface TerminalOptionListener {
   void configurationChange(final String name, Object oldValue, Object newValue);
}
