/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol;

import org.rgt.ByteArrayBuffer;
import org.rgt.Operation;

/**
 *
 * @author fabio_uggeri
 * @param <O> Operation served by protocol
 * @param <R> Request data
 * @param <S> Response data
 */
public abstract class AbstractProtocol<O extends Operation, R extends Request, S extends Response> implements Protocol<R, S> {

   private final O operation;

   private final ResponseCreator<S> responseCreator;


   public AbstractProtocol(final O operation, final ResponseCreator<S> responseCreator) {
      this.operation = operation;
      this.responseCreator = responseCreator;
   }

   protected abstract void serializeRequest(R request, ByteArrayBuffer buffer);

   protected abstract void deserializeResponse(ByteArrayBuffer buffer, S response);


   //*******************************************
   //* REQUEST
   //*******************************************
   protected void requestToBuffer(R request, ByteArrayBuffer buffer) {
      serializeRequest(request, buffer);
   }

   @Override
   public void putRequest(R request, ByteArrayBuffer buffer) {
      prepareRequest(buffer);
      requestToBuffer(request, buffer);
      finalizeBuffer(buffer);
   }

   protected boolean errorToBuffer(S response, ByteArrayBuffer buffer) {
      if (response.isError()) {
         buffer.putString(response.getMessage());
         return true;
      }
      return false;
   }

   //*******************************************
   //* RESPONSE
   //*******************************************
   protected void bufferToResponse(ByteArrayBuffer buffer, S response) {
      if (response.isError()) {
         response.setMessage(buffer.getString());
      } else {
         deserializeResponse(buffer, response);
      }
   }

   @Override
   public S getResponse(ByteArrayBuffer buffer) {
      final S response = responseCreator.create(buffer.getShort());
      bufferToResponse(buffer, response);
      return response;
   }

   private void prepareRequest(ByteArrayBuffer buffer) {
      buffer.clear();
      buffer.putInt(0);
      buffer.put(operation.getCode());
   }

   private void finalizeBuffer(ByteArrayBuffer buffer) {
      final int pos = buffer.position();
      buffer.flip();
      buffer.putInt(pos - 4);
      buffer.rewind();
   }

}
