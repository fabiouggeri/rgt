/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.client;

import ch.qos.logback.core.encoder.ByteArrayUtil;
import java.io.UnsupportedEncodingException;
import java.security.InvalidAlgorithmParameterException;
import java.security.InvalidKeyException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.KeySpec;
import java.util.Arrays;
import java.util.logging.Level;
import javax.crypto.BadPaddingException;
import javax.crypto.Cipher;
import javax.crypto.IllegalBlockSizeException;
import javax.crypto.NoSuchPaddingException;
import javax.crypto.SecretKey;
import javax.crypto.SecretKeyFactory;
import javax.crypto.spec.DESedeKeySpec;
import javax.crypto.spec.IvParameterSpec;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public class Security {

   private static final Logger LOG = LoggerFactory.getLogger(Security.class);

	private static final byte ADMIN_KEY[] = new byte[]{55, 48, 48, 51, 57, 52, 97, 99, 53, 102, 52, 48, 57, 53, 100, 50, 99, 50, 55,
      49, 99, 101, 99, 98, 101, 101, 55, 100, 54, 53, 52, 55, 98, 98, 57, 57, 50, 56, 53, 52, 55, 98, 57, 48, 49, 53, 55, 97, 56,
      98, 48, 101, 98, 49, 49, 53, 48, 97, 52, 49, 101, 49, 99, 57};

   private static final String DESEDE_ENCRYPTION_SCHEME = "DESede";

   public static final String ENCRYPTION_SCHEME = "DESede/CBC/NoPadding";

   private static Cipher cipher;

   private static SecretKey secretKey;

   private static IvParameterSpec ivBlockParam;

   private static Cipher cipher() {
      if (cipher == null) {
         try {
            final byte[] ivBlock;
            final MessageDigest md = MessageDigest.getInstance("SHA-256");
            final byte[] digestOfPassword = md.digest(ADMIN_KEY);
            final byte[] keyBytes;
            final KeySpec keySpec;
            final SecretKeyFactory keyFactory;

            keyBytes = Arrays.copyOf(digestOfPassword, 24);
            ivBlock = Arrays.copyOfRange(digestOfPassword, 24, 32);
            ivBlockParam = new IvParameterSpec(ivBlock);
            keySpec = new DESedeKeySpec(keyBytes);
            keyFactory = SecretKeyFactory.getInstance(DESEDE_ENCRYPTION_SCHEME);
            secretKey = keyFactory.generateSecret(keySpec);
            cipher = Cipher.getInstance(ENCRYPTION_SCHEME);
         } catch (NoSuchAlgorithmException | InvalidKeyException | NoSuchPaddingException | InvalidKeySpecException ex) {
            LOG.error("Error creating cipher", ex);
         }
      }
      return cipher;
   }

   private static byte[] padBuffer(byte[] buffer) {
      int resto = 8 - (buffer.length % 8);
      int padBufferLen = buffer.length + resto;
      byte padBuffer[];

      padBuffer = new byte[padBufferLen];
      System.arraycopy(buffer, 0, padBuffer, 0, buffer.length);
      /* Armazena o nro de bytes extras no ultimo byte do buffer */
      /* Preenche o restante dos bytes extras com 0xff */
      for (int i = resto; i > 0; i--) {
         padBuffer[padBufferLen - i] = (byte) (resto & 0xff);
      }
      return padBuffer;
   }

   public static String encrypt(final String text) {
      try {
         final byte[] encryptedBuffer;
         cipher().init(Cipher.ENCRYPT_MODE, secretKey, ivBlockParam);
         encryptedBuffer = cipher().doFinal(padBuffer(text.getBytes()));
         return "{DESede}" + ByteArrayUtil.toHexString(encryptedBuffer);
      } catch (IllegalBlockSizeException
              | BadPaddingException
              | InvalidKeyException
              | InvalidAlgorithmParameterException ex) {
         LOG.error("Error encrypting text", ex);
      }
      return text;
   }
}
