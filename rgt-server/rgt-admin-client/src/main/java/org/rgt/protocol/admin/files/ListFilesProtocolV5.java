/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.files;

import org.rgt.ByteArrayBuffer;
import static org.rgt.admin.AdminOperation.LIST_FILES;
import org.rgt.protocol.AdminProtocolRegister;

/**
 *
 * @author fabio_uggeri
 */
@AdminProtocolRegister(operation = LIST_FILES, fromVersion = 5)
public class ListFilesProtocolV5 extends ListFilesProtocol {

   @Override
   protected void deserializeResponse(ByteArrayBuffer buffer, ListFilesResponse response) {
      final int filesCount;
      response.folderPathname(buffer.getString());
      filesCount = buffer.getInt();
      for (int i = 0; i < filesCount; i++) {
         response.fileInfo(new FileInfo(buffer.getString(),
                 FileInfo.FileType.byValue(buffer.getByte()),
                 buffer.getLong(),
                 buffer.getDateTime().getTime(),
                 buffer.getDateTime().getTime()));
      }
   }
}
