/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.auth;

import ch.qos.logback.core.encoder.ByteArrayUtil;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.reflect.TypeToken;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileReader;
import java.io.FileWriter;
import java.io.IOException;
import java.io.Reader;
import java.io.UnsupportedEncodingException;
import java.io.Writer;
import java.lang.reflect.Type;
import java.security.InvalidKeyException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.KeySpec;
import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import javax.crypto.BadPaddingException;
import javax.crypto.Cipher;
import javax.crypto.IllegalBlockSizeException;
import javax.crypto.NoSuchPaddingException;
import javax.crypto.SecretKey;
import javax.crypto.SecretKeyFactory;
import javax.crypto.spec.DESedeKeySpec;
import org.rgt.admin.AdminResponseCode;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public class TerminalUserAuthenticator implements UserAuthenticator, UserRepository {

   private static final Logger LOG = LoggerFactory.getLogger(TerminalUserAuthenticator.class);

   private final Gson gson = new GsonBuilder().create();

   private static final String UNICODE_FORMAT = "UTF8";

   public static final String DESEDE_ENCRYPTION_SCHEME = "DESede";

   private static final byte DES_KEY[] = new byte[]{32, 42, 87, 61, 54, 29, 42, 12, 4, 95, 37, 84, 92};
   private Cipher cipher;
   private SecretKey secretKey;
   private Map<String, TerminalUser> users = new HashMap<>();

   public TerminalUserAuthenticator() throws AuthenticatorException {
      try {
         final MessageDigest md = MessageDigest.getInstance("md5");
         final byte[] digestOfPassword = md.digest(DES_KEY);
         final byte[] keyBytes = Arrays.copyOf(digestOfPassword, 24);
         final KeySpec keySpec;
         final SecretKeyFactory keyFactory;

         for (int j = 0, k = 16; j < 8;) {
            keyBytes[k++] = keyBytes[j++];
         }
         keySpec = new DESedeKeySpec(keyBytes);
         keyFactory = SecretKeyFactory.getInstance(DESEDE_ENCRYPTION_SCHEME);
         cipher = Cipher.getInstance(DESEDE_ENCRYPTION_SCHEME);
         secretKey = keyFactory.generateSecret(keySpec);
      } catch (NoSuchAlgorithmException | java.security.InvalidKeyException | NoSuchPaddingException | InvalidKeySpecException ex) {
         throw new AuthenticatorException(AdminResponseCode.AUTHENTICATOR_ERROR, ex);
      }
   }

   @Override
   public String encrypt(final String unencryptedString) throws AuthenticatorException {
      try {
         cipher.init(Cipher.ENCRYPT_MODE, secretKey);
         byte[] plainText = unencryptedString.getBytes(UNICODE_FORMAT);
         byte[] encryptedText = cipher.doFinal(plainText);
         return ByteArrayUtil.toHexString(encryptedText);
      } catch (InvalidKeyException | UnsupportedEncodingException | IllegalBlockSizeException | BadPaddingException ex) {
         throw new AuthenticatorException(AdminResponseCode.AUTHENTICATOR_ERROR, ex);
      }
   }

   @Override
   public String decrypt(final String encryptedString) throws AuthenticatorException {
      try {
         cipher.init(Cipher.DECRYPT_MODE, secretKey);
         byte[] encryptedText = ByteArrayUtil.hexStringToByteArray(encryptedString);
         byte[] decryptedText = cipher.doFinal(encryptedText);
         return new String(decryptedText);
      } catch (InvalidKeyException | IllegalBlockSizeException | BadPaddingException ex) {
         throw new AuthenticatorException(AdminResponseCode.AUTHENTICATOR_ERROR, ex);
      }
   }

   @Override
   public boolean authenticate(final String username, final String password) throws AuthenticatorException {
      final TerminalUser user = users.get(username);
      if (user != null) {
         return user.getPassword().equals(encrypt(password));
      }
      return false;
   }

   @Override
   public TerminalUser addUser(final String username, final String password) throws AuthenticatorException {
      final TerminalUser user = new TerminalUser(username, encrypt(password));
      if (addUser(user)) {
         return user;
      }
      return null;
   }

   @Override
   public boolean addUser(final TerminalUser user) throws AuthenticatorException {
      if (!users.containsKey(user.getUsername())) {
         users.put(user.getUsername(), user);
         return true;
      }
      return false;
   }

   @Override
   public TerminalUser removeUser(final String username) throws AuthenticatorException {
      return users.remove(username);
   }

   @Override
   public TerminalUser findUser(final String username) {
      return users.get(username);
   }

   @Override
   public Collection<TerminalUser> getUsers() {
      return Collections.unmodifiableCollection(users.values());
   }

   @Override
   public void clearUsers() throws AuthenticatorException {
      users.clear();
   }

   @Override
   public void save() throws AuthenticatorException {
      final File file = new File(System.getProperty("user.dir"), "users.json");
      try (Writer writer = new BufferedWriter(new FileWriter(file))) {
         writer.write(gson.toJson(users));
      } catch (IOException ex) {
         LOG.error("Error writing users file", ex);
         throw new AuthenticatorException(AdminResponseCode.AUTHENTICATOR_ERROR, "Error writing users file.", ex);
      }
   }

   @Override
   public void load() throws AuthenticatorException {
      final File file = new File(System.getProperty("user.dir"), "users.json");
      if (file.isFile()) {
         try (Reader reader = new BufferedReader(new FileReader(file))) {
            Type listType = new TypeToken<Map<String, TerminalUser>>(){}.getType();
            users = gson.fromJson(reader, listType);
         } catch (FileNotFoundException ex) {
            LOG.error("Users file not found", ex);
            throw new AuthenticatorException(AdminResponseCode.AUTHENTICATOR_ERROR, "Users file not found", ex);
         } catch (IOException ex) {
            LOG.error("Error reading users file", ex);
            throw new AuthenticatorException(AdminResponseCode.AUTHENTICATOR_ERROR, "Error reading users file", ex);
         }
      } else {
         LOG.warn("Users file not found");
      }
   }
}
