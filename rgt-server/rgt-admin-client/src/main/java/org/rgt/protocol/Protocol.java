/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol;

import java.util.Formatter;
import org.rgt.ByteArrayBuffer;

/**
 *
 * @author fabio_uggeri
 * @param <R>
 * @param <S>
 */
public interface Protocol<R extends Request, S extends Response> {

   final static Formatter FORMATTER = new Formatter();

   void putRequest(final R request, final ByteArrayBuffer buffer);

   default ByteArrayBuffer putRequest(final R data) {
      final ByteArrayBuffer buffer = new ByteArrayBuffer();
      putRequest(data, buffer);
      return buffer;
   }

   S getResponse(final ByteArrayBuffer buffer);

}
