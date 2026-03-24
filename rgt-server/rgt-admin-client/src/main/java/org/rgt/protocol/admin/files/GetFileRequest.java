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
public class GetFileRequest extends Request {

   private String pathFilename;

   public GetFileRequest(final String filePathname) {
      super(AdminOperation.GET_FILE);
      this.pathFilename = filePathname;
   }

   public GetFileRequest() {
      this(null);
   }

   public String pathFilename() {
      return pathFilename;
   }

   public GetFileRequest pathFilename(String pathFilename) {
      this.pathFilename = pathFilename;
      return this;
   }

}
