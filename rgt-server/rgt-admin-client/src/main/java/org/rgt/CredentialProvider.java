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
public interface CredentialProvider {

   Credential getCredential(final String id, final boolean newCredential);
   void registerCredential(final String id, final Credential credential);
   boolean isInteractive();
}
