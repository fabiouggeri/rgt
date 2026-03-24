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
public interface ProtocolProvider {

   <P extends Protocol<? extends Request, ? extends Response>> P getProtocol(Operation operation, short version)
           throws ProtocolErrorException;

}
