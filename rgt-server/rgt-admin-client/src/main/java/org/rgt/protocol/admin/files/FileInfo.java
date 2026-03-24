/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import java.io.File;
import java.nio.file.attribute.BasicFileAttributes;
import java.time.Instant;
import java.util.Calendar;
import java.util.Date;
import org.rgt.TerminalUtil;

/**
 *
 * @author fabio_uggeri
 */
public class FileInfo {

   public enum FileType {
      FILE((byte) 0),
      DIRECTORY((byte) 1);

      private final byte value;

      private FileType(byte value) {
         this.value = value;
      }

      public byte value() {
         return value;
      }

      public static FileType byValue(final byte value) {
         for (FileType ft : values()) {
            if (ft.value == value) {
               return ft;
            }
         }
         return FILE;
      }
   }

   private String name;
   private long length;
   private Date creationTime;
   private Date lastModificationTime;
   private FileType fileType;

   public FileInfo(final String name, final FileType fileType, final long length, final Date creationTime,
           final Date lastModificationTime) {
      this.name = name;
      this.length = length;
      this.creationTime = creationTime;
      this.lastModificationTime = lastModificationTime;
      this.fileType = fileType;
   }

   public FileInfo(final File file) {
      final BasicFileAttributes attr = TerminalUtil.getAttributes(file);
      this.name = file.getName();
      this.length = file.length();
      if (attr != null) {
         this.creationTime = new Date(attr.creationTime().toMillis());
         this.lastModificationTime = new Date(attr.lastModifiedTime().toMillis());
      } else {
         this.creationTime = new Date(Instant.EPOCH.toEpochMilli());
         this.lastModificationTime = this.creationTime;
      }
      this.fileType = file.isDirectory() ? FileType.DIRECTORY : FileType.FILE;
   }

   public FileInfo(final String fileName) {
      this(fileName, FileType.FILE, 0L, Calendar.getInstance().getTime(), Calendar.getInstance().getTime());
   }

   /**
    * @return the name
    */
   public String getName() {
      return name;
   }

   /**
    * @param name the name to set
    */
   public void setName(String name) {
      this.name = name;
   }

   /**
    * @return the length
    */
   public long getLength() {
      return length;
   }

   /**
    * @param length the length to set
    */
   public void setLength(long length) {
      this.length = length;
   }

   /**
    * @return the creationTime
    */
   public Date getCreationTime() {
      return creationTime;
   }

   /**
    * @param creationTime the creationTime to set
    */
   public void setCreationTime(Date creationTime) {
      this.creationTime = creationTime;
   }

   /**
    * @return the lastModificationTime
    */
   public Date getLastModificationTime() {
      return lastModificationTime;
   }

   /**
    * @param lastModificationTime the lastModificationTime to set
    */
   public void setLastModificationTime(Date lastModificationTime) {
      this.lastModificationTime = lastModificationTime;
   }

   public FileType getFileType() {
      return fileType;
   }

   public void setFileType(FileType fileType) {
      this.fileType = fileType;
   }

   public boolean isDirectory() {
      return this.fileType == FileType.DIRECTORY;
   }

   public boolean isFile() {
      return this.fileType == FileType.FILE;
   }

   @Override
   public String toString() {
      return name;
   }
}
