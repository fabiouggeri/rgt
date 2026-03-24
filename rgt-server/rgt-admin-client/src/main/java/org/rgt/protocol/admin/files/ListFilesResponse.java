/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template fileInfo, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import java.io.File;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import org.rgt.TerminalUtil;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.Response;

/**
 *
 * @author fabio_uggeri
 */
public class ListFilesResponse extends Response {

   private String folderPathname = "";

   private final ArrayList<FileInfo> filesInfo = new ArrayList<>();

   public ListFilesResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public ListFilesResponse(AdminResponseCode responseCode) {
      this(responseCode, null);
   }

   public ListFilesResponse(final String folderPathname) {
      this(AdminResponseCode.SUCCESS);
      this.folderPathname = folderPathname;
   }

   public ListFilesResponse(File folder, File files[]) {
      this(TerminalUtil.getAbsolutePath(folder));
      for (final File f : files) {
         filesInfo.add(new FileInfo(f));
      }
   }

   public String folderPathname() {
      return folderPathname;
   }

   public ListFilesResponse folderPathname(String folderPathname) {
      this.folderPathname = folderPathname;
      return this;
   }

   public ListFilesResponse fileInfo(final FileInfo file) {
      this.filesInfo.add(file);
      return this;
   }

   public ListFilesResponse filesInfo(final FileInfo... files) {
      Collections.addAll(this.filesInfo, files);
      return this;
   }

   public List<FileInfo> filesInfo() {
      return filesInfo;
   }
}
