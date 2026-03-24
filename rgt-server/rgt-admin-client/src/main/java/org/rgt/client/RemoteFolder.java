/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.client;

import java.util.ArrayList;
import java.util.List;
import org.rgt.protocol.admin.files.FileInfo;

/**
 *
 * @author fabio_uggeri
 */
public class RemoteFolder {

   private String pathName;

   private final List<FileInfo> files = new ArrayList<>();

   public RemoteFolder(String pathName) {
      this.pathName = pathName;
   }

   public String pathName() {
      return pathName;
   }

   public RemoteFolder pathName(String pathName) {
      this.pathName = pathName;
      return this;
   }

   public List<FileInfo> files() {
      return files;
   }

   public RemoteFolder files(List<FileInfo> files) {
      this.files.addAll(files);
      return this;
   }
}
