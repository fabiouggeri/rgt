/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template fileInfo, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import java.io.File;
import org.rgt.admin.AdminResponseCode;
import org.rgt.protocol.Response;

/**
 *
 * @author fabio_uggeri
 */
public class GetFileResponse extends Response {

   private FileInfo fileInfo;

   private byte data[];

   public GetFileResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public GetFileResponse(AdminResponseCode responseCode) {
      this(responseCode, null);
   }

   public GetFileResponse(final FileInfo fileInfo) {
      this(AdminResponseCode.SUCCESS, null);
      this.fileInfo = fileInfo;
   }

   public GetFileResponse(final File file) {
      this(new FileInfo(file));
   }

   public GetFileResponse fileInfo(final FileInfo fileInfo) {
      this.fileInfo = fileInfo;
      return this;
   }

   public GetFileResponse fileInfo(final File file) {
      this.fileInfo = new FileInfo(file);
      return this;
   }

   public FileInfo fileInfo() {
      return fileInfo;
   }

   public int dataSize() {
      return data != null ? data.length : 0;
   }

   public byte[] data() {
      return data;
   }

   public GetFileResponse data(byte[] data) {
      this.data = data;
      return this;
   }
}
