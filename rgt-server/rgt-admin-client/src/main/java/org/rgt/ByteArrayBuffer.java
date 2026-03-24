/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.Charset;
import java.util.Calendar;
import java.util.Date;

/**
 *
 * @author fabio_uggeri
 */
public class ByteArrayBuffer {

   private final static int DEFAULT_BUFFER_SIZE = 256;

   private static final int DATE_SIZE = Short.BYTES + Byte.BYTES + Byte.BYTES;

   private static final int HOUR_SIZE = Byte.BYTES + Byte.BYTES + Byte.BYTES;

   private ByteBuffer byteBuffer;
   private Charset charset;

   public ByteArrayBuffer(Charset charset, int bufferSize) {
      this.byteBuffer = ByteBuffer.allocate(bufferSize).order(ByteOrder.LITTLE_ENDIAN);
      this.charset = charset;
   }

   public ByteArrayBuffer(final ByteBuffer byteBuffer) {
      this.byteBuffer = byteBuffer;
      this.charset = Charset.forName("US-ASCII");
   }

   public ByteArrayBuffer(int bufferSize) {
      this(Charset.forName("IBM-850"), bufferSize);
   }

   public ByteArrayBuffer() {
      this(DEFAULT_BUFFER_SIZE);
   }

   public ByteArrayBuffer(Charset charset) {
      this(charset, DEFAULT_BUFFER_SIZE);
   }

   public ByteArrayBuffer(Charset charset, byte byteArray[]) {
      this.charset = charset;
      this.byteBuffer = ByteBuffer.wrap(byteArray).order(ByteOrder.LITTLE_ENDIAN);
   }

   public ByteArrayBuffer(String charset, byte byteArray[]) {
      this(Charset.forName(charset), byteArray);
   }

   public ByteArrayBuffer(byte byteArray[]) {
      this("US-ASCII", byteArray);
   }

   public void setCharset(final Charset charset) {
      this.charset = charset;
   }

   public String getString() {
      final String result;
      final int len = byteBuffer.getInt();
      final byte buffer[] = new byte[len];
      byteBuffer.get(buffer);
      result = new String(buffer, charset);
      return result;
   }

   public ByteArrayBuffer get(byte dst[]) {
      final int len = byteBuffer.getInt();
      byteBuffer.get(dst, 0, len);
      return this;
   }

   public ByteArrayBuffer get(byte dst[], int offset, int length) {
      final int len = byteBuffer.getInt();
      if (length > len) {
         throw new IndexOutOfBoundsException("Array in buffer is smaller than given length."
                 + " Array len: " + len + " Required len: " + length);
      }
      byteBuffer.get(dst, offset, length);
      return this;
   }

   public ByteArrayBuffer getBytes(byte dst[], int length) {
      byteBuffer.get(dst, 0, length);
      return this;
   }

   public ByteArrayBuffer getBytes(byte dst[]) {
      byteBuffer.get(dst);
      return this;
   }

   public byte[] getArray() {
      final byte array[] = new byte[byteBuffer.getInt()];
      byteBuffer.get(array);
      return array;
   }

   public boolean getBoolean() {
      return byteBuffer.get() != 0;
   }

   public byte getByte(int index) {
      return byteBuffer.get(index);
   }

   public byte getByte() {
      return byteBuffer.get();
   }

   public short getShort() {
      return byteBuffer.getShort();
   }

   public int getInt() {
      return byteBuffer.getInt();
   }

   public long getLong() {
      return byteBuffer.getLong();
   }

   public Double getDouble() {
      return byteBuffer.getDouble();
   }

   public Float getFloat() {
      return byteBuffer.getFloat();
   }

   public Calendar getDate() {
      Calendar calendar = Calendar.getInstance();
      calendar.clear();
      calendar.set(byteBuffer.getShort(), byteBuffer.get() - 1, byteBuffer.get());
      return calendar;
   }

   public Calendar getDateTime() {
      Calendar calendar = Calendar.getInstance();
      calendar.clear();
      calendar.set(byteBuffer.getShort(), byteBuffer.get() - 1, byteBuffer.get(),
                   byteBuffer.get(), byteBuffer.get(), byteBuffer.get());
      return calendar;
   }

   private void ensureCapacity(final int increase) {
      int newCapacity;
      int minLen;
      if (increase <= 0) {
         return;
      }
      minLen = byteBuffer.position() + increase;
      if (minLen <= byteBuffer.limit()) {
         return;
      }
      if (minLen <= byteBuffer.capacity()) {
         byteBuffer.limit(minLen);
         return;
      }
      newCapacity = (byteBuffer.capacity() * 3) / 2 + 1;
      if (newCapacity < minLen) {
         newCapacity = minLen;
      }
      byteBuffer.flip();
      byteBuffer = ByteBuffer.allocate(newCapacity).order(ByteOrder.LITTLE_ENDIAN).put(byteBuffer);
   }

   public ByteArrayBuffer putString(final String value) {
      final byte array[];
      if (!isEmptyString(value)) {
         array = value.getBytes(charset);
         ensureCapacity(array.length + Integer.BYTES);
         byteBuffer.putInt(array.length);
         byteBuffer.put(array);
      } else {
         ensureCapacity(Integer.BYTES);
         byteBuffer.putInt(0);
      }
      return this;
   }

   public ByteArrayBuffer putObjects(Object... values) {
      for (Object o : values) {
         if (Integer.class.isInstance(o)) {
            putInt((Integer) o);
         } else if (Long.class.isInstance(o)) {
            putLong((Long) o);
         } else if (String.class.isInstance(o)) {
            putString((String) o);
         } else if (Short.class.isInstance(o)) {
            putShort((Short) o);
         } else if (Byte.class.isInstance(o)) {
            put((Byte) o);
         } else if (Boolean.class.isInstance(o)) {
            putBoolean((Boolean) o);
         } else if (Date.class.isInstance(o)) {
            putDate((Date) o);
         } else if (Calendar.class.isInstance(o)) {
            putDate(((Calendar) o).getTime());
         } else if (Double.class.isInstance(o)) {
            putDouble((Double) o);
         } else if (Float.class.isInstance(o)) {
            putFloat((Float) o);
         } else {
            putString(o.toString());
         }
      }
      return this;
   }

   public ByteArrayBuffer putBytes(final byte value[]) {
      ensureCapacity(value.length);
      byteBuffer.put(value);
      return this;
   }

   public ByteArrayBuffer putBytes(final byte value[], int offset, int length) {
      final int len = length <= value.length - offset ? length : value.length - offset;
      ensureCapacity(len);
      byteBuffer.put(value, offset, len);
      return this;
   }

   public ByteArrayBuffer put(final byte value[]) {
      ensureCapacity(Integer.BYTES + value.length);
      byteBuffer.putInt(value.length);
      byteBuffer.put(value);
      return this;
   }

   public ByteArrayBuffer put(final byte value[], int offset, int length) {
      final int len = length <= value.length - offset ? length : value.length - offset;
      ensureCapacity(Integer.BYTES + len);
      byteBuffer.putInt(len);
      byteBuffer.put(value, offset, len);
      return this;
   }

   public ByteArrayBuffer put(final byte value) {
      ensureCapacity(Byte.BYTES);
      byteBuffer.put(value);
      return this;
   }

   public ByteArrayBuffer putBoolean(final boolean value) {
      byte boolValue = (byte) (value ? 1 : 0);
      ensureCapacity(Byte.BYTES);
      byteBuffer.put(boolValue);
      return this;
   }

   public ByteArrayBuffer putShort(final short value) {
      ensureCapacity(Short.BYTES);
      byteBuffer.putShort(value);
      return this;
   }

   public ByteArrayBuffer putInt(final int value) {
      ensureCapacity(Integer.BYTES);
      byteBuffer.putInt(value);
      return this;
   }

   public ByteArrayBuffer putLong(final long value) {
      ensureCapacity(Long.BYTES);
      byteBuffer.putLong(value);
      return this;
   }

   public ByteArrayBuffer putDouble(final double value) {
      ensureCapacity(Double.BYTES);
      byteBuffer.putDouble(value);
      return this;
   }

   public ByteArrayBuffer putFloat(final float value) {
      ensureCapacity(Float.BYTES);
      byteBuffer.putFloat(value);
      return this;
   }

   public ByteArrayBuffer putDate(Calendar date) {
      ensureCapacity(DATE_SIZE);
      byteBuffer.putShort((short) date.get(Calendar.YEAR));
      byteBuffer.put((byte) (date.get(Calendar.MONTH) + 1));
      byteBuffer.put((byte) date.get(Calendar.DAY_OF_MONTH));
      return this;
   }

   public ByteArrayBuffer putDate(Date date) {
      final Calendar cal = Calendar.getInstance();
      cal.setTime(date);
      return putDate(cal);
   }

   public ByteArrayBuffer putDateTime(Calendar date) {
      ensureCapacity(DATE_SIZE + HOUR_SIZE);
      byteBuffer.putShort((short) date.get(Calendar.YEAR));
      byteBuffer.put((byte) (date.get(Calendar.MONTH) + 1));
      byteBuffer.put((byte) date.get(Calendar.DAY_OF_MONTH));
      byteBuffer.put((byte) date.get(Calendar.HOUR_OF_DAY));
      byteBuffer.put((byte) date.get(Calendar.MINUTE));
      byteBuffer.put((byte) date.get(Calendar.SECOND));
      return this;
   }

   public ByteArrayBuffer putDateTime(Date date) {
      final Calendar cal = Calendar.getInstance();
      cal.setTime(date);
      return putDateTime(cal);
   }

   public ByteArrayBuffer put(ByteBuffer packetBuffer) {
      ensureCapacity(packetBuffer.remaining());
      byteBuffer.put(packetBuffer);
      return this;
   }

   public ByteArrayBuffer flip() {
      byteBuffer.flip();
      return this;
   }

   public ByteArrayBuffer rewind() {
      byteBuffer.rewind();
      return this;
   }

   public ByteArrayBuffer compact() {
      byteBuffer.compact();
      return this;
   }

   public byte[] backArray() {
      return byteBuffer.array();
   }

   public ByteArrayBuffer clear() {
      byteBuffer.clear();
      return this;
   }

   public int position() {
      return byteBuffer.position();
   }

   public void setPosition(int newPos) {
      byteBuffer.position(newPos);
   }

   public int limit() {
      return byteBuffer.limit();
   }

   public ByteArrayBuffer limit(int newLimit) {
      if (newLimit > byteBuffer.limit()) {
         ensureCapacity(newLimit);
      }
      byteBuffer.limit(newLimit);
      return this;
   }

   public int remaining() {
      return byteBuffer.remaining();
   }

   public ByteArrayBuffer slice() {
      return new ByteArrayBuffer(byteBuffer.slice().order(ByteOrder.LITTLE_ENDIAN).array());
   }

   private boolean isEmptyString(String value) {
      return value == null || value.isEmpty();
   }

   public ByteBuffer getByteBuffer() {
      return byteBuffer;
   }

   public int capacity() {
      return byteBuffer.capacity();
   }

   public Charset getCharset() {
      return charset;
   }

   @Override
   public String toString() {
      return byteBuffer.toString();
   }

   public boolean hasRemaining() {
      return byteBuffer.hasRemaining();
   }
}
