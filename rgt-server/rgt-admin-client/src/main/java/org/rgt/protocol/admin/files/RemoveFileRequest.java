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
public class RemoveFileRequest extends Request {

   private String remotePathname;
   private String filename;

   public RemoveFileRequest() {
      super(AdminOperation.REMOVE_FILE);
   }
   
   public RemoveFileRequest(final String remotePathname, final String filename) {
      super(AdminOperation.REMOVE_FILE);
      this.remotePathname = remotePathname;
      this.filename = filename;
   }

   /**
    * @return the remotePathname
    */
   public String remotePathname() {
      return remotePathname;
   }

   /**
    * @param remotePathname the remotePathname to set
    * @return this
    */
   public RemoveFileRequest remotePathname(String remotePathname) {
      this.remotePathname = remotePathname;
      return this;
   }

   /**
    * @return the filename
    */
   public String filename() {
      return filename;
   }

   /**
    * @param filename the filename to set
    * @return this
    */
   public RemoveFileRequest filename(String filename) {
      this.filename = filename;
      return this;
   }
}

