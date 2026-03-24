/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import java.util.Date;
import org.rgt.admin.AdminOperation;
import org.rgt.protocol.Request;

/**
 *
 * @author fabio_uggeri
 */
public class PutFileRequest extends Request {

   private String filePathname;

   private Date creationTime;

   private Date lastModificationTime;

   private long fileSize;

   private boolean force;

   private byte data[];

   public PutFileRequest() {
      this(null);
   }

   public PutFileRequest(final String filePathname) {
      super(AdminOperation.PUT_FILE);
      this.filePathname = filePathname;
   }

   public String filePathname() {
      return filePathname;
   }

   public PutFileRequest filePathname(String filePathname) {
      this.filePathname = filePathname;
      return this;
   }

   public Date creationTime() {
      return creationTime;
   }

   public PutFileRequest creationTime(Date creationTime) {
      this.creationTime = creationTime;
      return this;
   }

   public Date lastModificationTime() {
      return lastModificationTime;
   }

   public PutFileRequest lastModificationTime(Date lastModificationTime) {
      this.lastModificationTime = lastModificationTime;
      return this;
   }

   public boolean force() {
      return force;
   }

   public PutFileRequest force(boolean force) {
      this.force = force;
      return this;
   }

   public long fileSize() {
      return fileSize;
   }

   public PutFileRequest fileSize(long fileSize) {
      this.fileSize = fileSize;
      return this;
   }

   public int dataSize() {
      return data != null ? data.length : 0;
   }

   public byte[] data() {
      return data;
   }

   public PutFileRequest data(byte[] data) {
      this.data = data;
      return this;
   }

}
