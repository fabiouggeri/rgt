/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import org.rgt.admin.AdminOperation;
import org.rgt.protocol.Request;

/**
 *
 * @author fabio_uggeri
 */
public class ListFilesRequest extends Request {

   private String path;

   public ListFilesRequest(final String path) {
      super(AdminOperation.LIST_FILES);
      this.path = path;
   }

   public ListFilesRequest() {
      this(null);
   }

   public String path() {
      return path;
   }

   public ListFilesRequest path(String path) {
      this.path = path;
      return this;
   }

}
