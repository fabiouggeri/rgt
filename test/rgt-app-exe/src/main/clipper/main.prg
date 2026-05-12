#IfDef __HBR__
   #Include "hbgtinfo.ch"
#EndIf
#Include "setcurs.ch"
#include "inkey.ch"
#include "fileio.ch"

#Define COR_PADRAO_TELA "G/B+"

#Define COR_TITULO "N/W"

#Define COR_SHELL "W/N"

#Define COR_CAMPO "B/W"

#IfDef __LNX__
ANNOUNCE HB_GT_TRM

REQUEST HB_GT_NUL
#EndIf

REQUEST RGT_INIT

Request NetName
Request SIMULA_INTERACAO
Request INTERROMPE_SIMULACAO_INTERACAO
Request FERASE

/**
 * main
 *
 * @param cPar1 Parametro string.
 * @param nPar2 Parametro numeric.
 *
 * @author fabio_uggeri
 * @since 26/11/2018
 * @version
 */
function main( cRepeticoes, cTamanhoArq, cIntervaloArquivo, cIntervaloDigitacao )
   local cRetorno
   local nLoop
   local nLin
   local nCol
   local nKey
   local cKeys := ""
   local nTempoInicio
   local nRepeticoes := 0
   local nIntervaloArquivo := 0
   local nTamanhoArq := 0
   local cPid := AllTrim( Str( F_GetPid() ) )
   local cPathApp
   local cPathTe
   local cArqOri1
   local cArqOri2
   local cArqDest
   local nError
   local nIntervaloDigitacao
   local nIniRpc
   local nTempoRpc := 0
   local cRpcResult
   local nIniTransf
   local nTempoTransf := 0
   local nTransfCount := 0
   local nTeclasCount := 0
   local cUsuario := "USUARIO_NAO_IDENTIFICADO"
   local cSenha := ""
   local cNovaSenha := ""
   local nTempoIni

#IfDef __HBR__
   REQUEST HB_CODEPAGE_PT850
   hb_cdpSelect( "PT850" )
   hb_setTermCP( "PT850" )
   HB_GtInfo( HB_GTI_MOUSESTATUS, .F. )
#EndIf

   //Rgt_LogLevel(5)

   If HB_IsString( cRepeticoes )
      nRepeticoes := Val( cRepeticoes )
   EndIf
   If nRepeticoes == 0
      nRepeticoes := 1000
   EndIf

   If HB_IsString( cIntervaloArquivo )
      nIntervaloArquivo := Val( cIntervaloArquivo )
   EndIf
   If nIntervaloArquivo == 0
      nIntervaloArquivo := Int( nRepeticoes / 100 )
   EndIf

   If HB_IsString( cTamanhoArq )
      nTamanhoArq := Val( cTamanhoArq )
   EndIf
   If nTamanhoArq == 0
      nTamanhoArq := 30
   EndIf

   If HB_IsString( cIntervaloDigitacao )
      nIntervaloDigitacao := Val( cIntervaloDigitacao )
   EndIf
   If nIntervaloDigitacao == 0
      nIntervaloDigitacao := 30
   EndIf

   cPathApp := AllTrim( GetEnv( "TEMP" ) )
   If Empty( cPathApp )
#IfDef __LNX__
      cPathApp := "/tmp/"
#Else
      cPathApp := "c:\temp\"
#EndIf
   ElseIf ! Right( cPathApp, 1 ) == HB_OSPathSeparator()
      cPathApp += HB_OSPathSeparator()
   EndIf
   
   // rgt_UpdateInterval( 250 )   
   // Test memory health check
   // cUsuario := Space( 21474836480 )
   // Alert("Press ANY KEY")
   // ? Left( cUsuario, 10 )
   
   // Abre o arquivo de LOG
   OpenLog( cPathApp + "teste_terminal.log" )

   cPathTe := AllTrim( rgt_ExecRemote( "GETENV", "TEMP" ) )

   If Empty( cPathTe )
#IfDef __LNX__
      cPathTe := "/tmp/"
#Else
      cPathTe := "c:\temp\"
#EndIf
   ElseIf ! Right( cPathTe, 1 ) == HB_OSPathSeparator()
      cPathTe += HB_OSPathSeparator()
   EndIf

   cArqOri1 := cPathApp + "testeput1_" + cPid + ".txt"
   cArqOri2 := cPathApp + "testeput2_" + cPid + ".txt"
   cArqDest := cPathTe + "testeget_" + cPid + ".txt"

   SetMode(25,80)

   // SolicitaCredencial( 10, 12, @cUsuario, @cSenha, @cNovaSenha, "Login", .T. )

   CriaArquivo( cArqOri1, nTamanhoArq )
   nTempoInicio := Seconds()
   rgt_ExecRemote( "SIMULA_INTERACAO", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", nIntervaloDigitacao )
   For nLoop := 1 To nRepeticoes
      nIniRpc := Seconds()
      cRpcResult := rgt_ExecRemote( "Lower", "TRMLOWER(" + AllTrim(Chr( 65 + nLoop % 25)) + ")")
      nTempoRpc += Seconds() - nIniRpc
      QOut( nLoop, "-", Chr(65 + nLoop % 25), cRpcResult )
      nKey := Inkey()
      If nKey == K_ESC
         Exit
      ElseIf nKey != 0
         QQout( "key", "=", Chr( nKey ) )
         cKeys += Chr( nKey )
         ++nTeclasCount
      EndIf
      If nLoop % nIntervaloArquivo == 0
         nIniTransf := Seconds()
         nError := rgt_PutFile( cArqOri1, cArqDest )
         If nError != 0
            ? "Erro rgt_PutFile", nError
         EndIf
         FErase( cArqOri2 )
         nError := rgt_GetFile( cArqDest, cArqOri2 )
         nTempoTransf += Seconds() - nIniTransf
         ++nTransfCount
         If nError == 0
            ComparaArquivos( cArqOri1, cArqOri2 )
         Else
            ? "Erro rgt_GetFile", nError
         EndIf
      EndIf
   Next
   rgt_ExecRemote( "INTERROMPE_SIMULACAO_INTERACAO" )
   FErase( cArqOri1 )
   FErase( cArqOri2 )
   rgt_ExecRemote( "FERASE", cArqDest )

   LogInfo( "PID: ", cPid, Chr(13) + Chr(10),;
            "   Transferencias: ", nTransfCount, " Tempo : ", nTempoTransf, Chr(13) + Chr(10),;
            "   RPCs: ", nLoop - 1, " Tempo: ", nTempoRpc, Chr(13) + Chr(10),;
            "   Teclas recebidas: ", nTeclasCount, Chr(13) + Chr(10),;
            "   Tempo total: ", Seconds() - nTempoInicio)

   CloseLog()

Return Nil


Static Function CriaArquivo( cArquivo, nTamanho )
   Local nHandle
   Local cBuffer

   nHandle := FCreate( cArquivo, FC_NORMAL )
   If nHandle < 0
      ? "Nao foi possivel criar o arquivo para teste de envio e recebimento"
      QUIT
   EndIf

   cBuffer := Replicate( "0123456789ABCDEF", 64 )
   While nTamanho > 0
      FWrite( nHandle, cBuffer )
      --nTamanho
   EndDo

   FClose( nHandle )
Return Nil

Static Function ComparaArquivos( cArq1, cArq2 )
   Local nHandle1
   Local nHandle2
   Local cBuffer1 := Space( 1024 )
   Local cBuffer2 := Space( 1024 )
   Local nRead1 := 1
   Local nRead2 := 1
   Local nTotal1 := 0
   Local nTotal2 := 0

   If ! File( cArq1 )
      ? "Arquivo", cArq1, "nao econtrado"
      Quit
   EndIf

   If ! File( cArq2 )
      ? "Arquivo", cArq2, "nao econtrado"
      Quit
   EndIf

   nHandle1 := FOPEN( cArq1 )
   If FError() != 0
       ? "Erro", FError(), "ao abrir o arquivo", cArq1
       Quit
   EndIf

   nHandle2 := FOPEN( cArq2 )
   If FError() != 0
       FClose( nHandle1 )
       ? "Erro", FError(), "ao abrir o arquivo", cArq2
       Quit
   EndIf

   While nRead1 > 0 .And. nRead2 > 0
      nRead1 := FRead( nHandle1, @cBuffer1, 1024 )
      nRead2 := FRead( nHandle2, @cBuffer2, 1024 )
      nTotal1 += nRead1
      nTotal2 += nRead2
      If nRead1 == nRead2
         If ! cBuffer1 == cBuffer2
            ? "Arquivo", cArq1, "diferente de", cArq2, "(", cBuffer1, "<>", cBuffer2,")", Chr(13) + Chr(10)
            ? "Total lido arquivo", cArq1, ":", nTotal1, Chr(13) + Chr(10)
            ? "Total lido arquivo", cArq2, ":", nTotal2, Chr(13) + Chr(10)
            nRead1 := 0
         EndIf
      Else
         ? "Arquivo", cArq1, "com tamanho diferente de", cArq2, "(", nRead1, "<>", nRead2,")", Chr(13) + Chr(10)
         ? "Total lido arquivo", cArq1, ":", nTotal1, Chr(13) + Chr(10)
         ? "Total lido arquivo", cArq2, ":", nTotal2, Chr(13) + Chr(10)
         nRead1 := 0
      EndIf
   EndDo

   FClose( nHandle1 )
   FClose( nHandle2 )

Return Nil


/*--------------------------------------------------------------------------------------------------
  Funcao ........: SolicitaCredencial
  Objetivo ......: Exibe tela de login para o usuario, solicitando seu login e senha
  Parametros ....:
    Nome                 Descricao
    -------------------- ---------------------------------------------------------------------------
    <nLin>               Linha onde sera exibida a janela
    <nCol>               Coluna onde sera exibida a janela
    <@cUsuario>          Retorna o login do usuario
    <@cSenha>            Retorna a senha do usuario
    <@cNovaSenha>        Retorna a nova senha do usuario no caso de usuario ter solicitado a troca
    <cTitulo>            Titulo a ser exibido na janela de solicitacao da credencial
    <lAlteraSenha>       Indica se permitir a alteracao de senha

  Retorno .......:
    Valor                Descricao
    -------------------- ---------------------------------------------------------------------------
    <.T.>                Entrada de credenciais confirmada pelo usuario
    <.F.>                Entrada de credenciais cancelada pelo usuario

  Historico .....:
    Nome                 Versao   Data       Descricao
    -------------------- -------- ---------- -------------------------------------------------------
    <Fabio>              1.38     04/05/2009 Criacao da rotina.
  Observacoes ...:
--------------------------------------------------------------------------------------------------*/
Static Function SolicitaCredencial( nLin, nCol, cUsuario, cSenha, cNovaSenha, cTitulo, lAlteraSenha )
   Local lSucesso := .T.
   Local cTela := SaveScreen( nLin, nCol, nLin + 6, nCol + 55 ) // Salva uma area maior que a janela para o caso de solicitacao de troca de senha

   TelaAutenticacao( nLin, nCol, cTitulo, lAlteraSenha )

   EditaCredencial( nLin, nCol, @cUsuario, @cSenha, @cNovaSenha, lAlteraSenha )

   If LastKey() == K_ESC .Or. Empty( cUsuario ) .Or. Empty( cSenha )
      lSucesso := .F.
   EndIf

   RestScreen( nLin, nCol, nLin + 6, nCol + 55, cTela )

Return lSucesso

/*--------------------------------------------------------------------------------------------------
  Funcao ........: Infra_SolicitaTrocaSenha
  Objetivo ......: Solicita as informacoes necessarias para troca de senha
  Parametros ....:
    Nome                 Descricao
    -------------------- ---------------------------------------------------------------------------
    <nLin>               Retorna o login do usuario
    <nCol>               Retorna a senha do usuario
    <cUsuario>           Usuario para o qual se esta trocando a senha
    <cSenhaAtual>        Senha atual do usuario
    <@cNovaSenha>        Nova senha do usuario
    <lDesenhaJanela>     Informa se deve desenhar a janela de troca de senha

  Retorno .......:
    Valor                Descricao
    -------------------- ---------------------------------------------------------------------------
    <.T.>                Troca de senha efetuada
    <.F.>                Troca de senha nao efetuada

  Historico .....:
    Nome                 Versao   Data       Descricao
    -------------------- -------- ---------- -------------------------------------------------------
    <Fabio>              1.38     25/05/2009 Criacao da rotina.
  Observacoes ...:
--------------------------------------------------------------------------------------------------*/
Function Infra_SolicitaTrocaSenha( nLin, nCol, cUsuario, cSenhaAtual, cNovaSenha, lDesenhaJanela )
   Local lSucesso := .T.

   If ValType( lDesenhaJanela ) == "L" .And. lDesenhaJanela
      TelaTrocaSenha( nLin, nCol, .T. )
      EditaTrocaSenha( nLin, nCol, cUsuario, cSenhaAtual, @cNovaSenha )
   Else
      EditaTrocaSenha( nLin, nCol, Nil, cSenhaAtual, @cNovaSenha )
   EndIf

   If LastKey() == K_ESC .Or. Empty( cNovaSenha )
      lSucesso := .F.
      cNovaSenha := ""
   EndIf

Return lSucesso



/*--------------------------------------------------------------------------------------------------
  Funcao ........: TelaAutenticacao
  Objetivo ......: Montar a tela de autenticacao de usuario
  Parametros ....:
    Nome                 Descricao
    -------------------- ---------------------------------------------------------------------------
    <nLin>               Linha base da janela de login
    <nCol>               Coluna base da janela de login
    <cTitulo>            Titulo a ser exibido na janela
    <lAlteraSenha>       Indica se deve exibir a mensagem de alteracao de senha

  Historico .....:
    Nome                 Versao   Data       Descricao
    -------------------- -------- ---------- -------------------------------------------------------
    <Fabio>              1.38     05/05/2009 Criacao da rotina.
  Observacoes ...:
--------------------------------------------------------------------------------------------------*/
Static Function TelaAutenticacao( nLin, nCol, cTitulo, lAlteraSenha )
   Local cBarraTitulo

   cTitulo := " " + AllTrim( Left( cTitulo, 49 ) ) + " "
   cBarraTitulo := Space( Len( cTitulo ) ) + "+" + Replicate( "-", 51 - Len( cTitulo ) )

   DispOutAt( nLin, nCol, "+-¦" + cBarraTitulo + "+", COR_PADRAO_TELA )
   DispOutAt( nLin + 1, nCol, "¦ Usuário ..........: [                              ] ¦", COR_PADRAO_TELA )
   DispOutAt( nLin + 2, nCol, "¦ Senha ............: [                              ] ¦", COR_PADRAO_TELA )
   If lAlteraSenha
      DispOutAt( nLin + 3, nCol, "+-[ <ESC> para retornar <F2> Troca senha ]-------------+", COR_PADRAO_TELA )
   Else
      DispOutAt( nLin + 3, nCol, "+-[ <ESC> para retornar ]------------------------------+", COR_PADRAO_TELA )
   EndIf
   DispOutAt( nLin, nCol + 03, cTitulo, COR_TITULO )
   DispOutAt( nLin + 1, nCol + 23, Space( 30 ), COR_CAMPO )
   DispOutAt( nLin + 2, nCol + 23, Space( 30 ), COR_CAMPO )
Return Nil


/*--------------------------------------------------------------------------------------------------
  Funcao ........: TelaTrocaSenha
  Objetivo ......: Montar a tela de troca de senha
  Parametros ....:
    Nome                 Descricao
    -------------------- ---------------------------------------------------------------------------
    <nLin>               Linha base da janela de login
    <nCol>               Coluna base da janela de login
    [lJanela]            Indica se eh uma janela ou extensao de uma janela

  Historico .....:
    Nome                 Versao   Data       Descricao
    -------------------- -------- ---------- -------------------------------------------------------
    <Fabio>              1.38     25/05/2009 Criacao da rotina.
  Observacoes ...:
--------------------------------------------------------------------------------------------------*/
Static Function TelaTrocaSenha( nLin, nCol, lJanela )
   If ValType( lJanela ) == "L" .And. lJanela
      DispOutAt( nLin++, nCol, "+-¦ Troca de senha +-----------------------------------+", COR_PADRAO_TELA )
      DispOutAt( nLin++, nCol, "¦ Usuario ..........: [                              ] ¦", COR_PADRAO_TELA )
   Else
      DispOutAt( nLin++, nCol, "+-¦ Troca de senha +-----------------------------------¦", COR_PADRAO_TELA )
   EndIf
   DispOutAt( nLin, nCol, "¦ Nova senha .......: [                              ] ¦", COR_PADRAO_TELA )
   DispOutAt( nLin++, nCol + 23, Space( 30 ), COR_CAMPO )
   DispOutAt( nLin, nCol, "¦ Confirmacao senha : [                              ] ¦", COR_PADRAO_TELA )
   DispOutAt( nLin++, nCol + 23, Space( 30 ), COR_CAMPO )
   DispOutAt( nLin++, nCol, "+-[ <ESC> para retornar ]------------------------------+", COR_PADRAO_TELA )
Return Nil


/*--------------------------------------------------------------------------------------------------
  Funcao ........: EditaCredencial
  Objetivo ......: Monta os gets para edicao da credencial do usuario
  Pre-Requisitos :
  Parametros ....:
    Nome                 Descricao
    -------------------- ---------------------------------------------------------------------------
    <nLin>               Linha base da janela de login
    <nCol>               Coluna base da janela de login
    <@cUsuario>          Variavel onde sera retornado o usuario informado
    <@cSenha>            Variavel onde sera retornada a senha informada
    <@cNovaSenha>        Variavel onde sera retornada a nova senha no caso de usuario ter solicitado
                         a troca
    <lAlteraSenha>       Indica se eh permitida a alteracao de senha

  Historico .....:
    Nome                 Versao   Data       Descricao
    -------------------- -------- ---------- -------------------------------------------------------
    <Fabio>              1.38     05/05/2009 Criacao da rotina.
    <Borba>              3.10.0   27/09/2018 Tratamento para logar também com usuário LDAP
  Observacoes ...:
--------------------------------------------------------------------------------------------------*/
Static Function EditaCredencial( nLin, nCol, cUsuario, cSenha, cNovaSenha, lAlteraSenha )
   Local oGet
   Local aGets := {}
   Local bBlocoF2
   Local lTrocaSenha := .F.
   Local cTela := SaveScreen( nLin + 3, nCol, nLin + 6, nCol + 55 )
   Local lContinua := .T.
   Local cCorAnt

   cSenha := IIf( Empty( cSenha ), Space( 32 ), cSenha )

   While lContinua
      lContinua := .F.

      cUsuario := PadR( cUsuario, 50 )
      cSenha := PadR( cSenha, 32 )

      /* Bloco para ativar ou desativar a troca de senha */
      If lAlteraSenha
         bBlocoF2 := SetKey( K_F2, { || lTrocaSenha := ! lTrocaSenha,;
                                     IIf( lTrocaSenha, TelaTrocaSenha( nLin + 3, nCol ),;
                                          RestScreen( nLin + 3, nCol, nLin + 6, nCol + 55, cTela ) ) } )
      EndIf

      // Solicita o login do usuario
      oGet := GetNew( nLin + 1, nCol + 23, { | x | IIf( x == NIL, cUsuario, cUsuario := x ) }, "cUsuario", "@K!S30" )
      oGet:PreBlock := { || DispOutAt( 24, 1, "Por favor, informe seu usuario.                                             ", .T. ),;
                            .T. }
      oGet:PostBlock := { || ! Empty( cUsuario ) }
      AAdd( aGets, oGet )

      // Solicita a senha do usuario
      oGet := GetNew( nLin + 2, nCol + 23, { | x | IIf( x == NIL, cSenha, cSenha := x ) }, "cSenha", "@KS30" )
      oGet:PreBlock := { || DispOutAt( 24, 1, "Por favor, informe sua senha.                                                ", .T. ),;
                            .T. }
      oGet:PostBlock := { || ! Empty( cSenha ) }
      AAdd( aGets, oGet )

      SetCursor( SC_NORMAL )
      ReadModal( aGets )
      SetCursor( SC_NONE )

      If lAlteraSenha
         SetKey( K_F2, bBlocoF2 )
      EndIf

      If Lastkey() != K_ESC .And. lTrocaSenha
         Infra_SolicitaTrocaSenha( nLin + 3, nCol, cUsuario, cSenha, @cNovaSenha, .F. )
         If LastKey() == K_UP
            lContinua := .T.
         EndIf
      EndIf
   EndDo
   cUsuario := AllTrim( cUsuario )
   cSenha := AllTrim( cSenha )

Return Nil

/*--------------------------------------------------------------------------------------------------
  Funcao ........: EditaTrocaSenha
  Objetivo ......: Monta os gets para troca de senha
  Pre-Requisitos :
  Parametros ....:
    Nome                 Descricao
    -------------------- ---------------------------------------------------------------------------
    <nLin>               Linha base da janela de login
    <nCol>               Coluna base da janela de login
    [cUsuario]           Usuario para o qual se esta trocando a senha. Somente para exibicao. Caso
                         nao seja passado, nao exibe o campo correspondente.
    <cSenhaAtual>        Senha atual do usuario
    <@cNovaSenha>        Variavel onde sera retornada a nova senha

  Historico .....:
    Nome                 Versao   Data       Descricao
    -------------------- -------- ---------- -------------------------------------------------------
    <Fabio>              1.38     25/05/2009 Criacao da rotina.
  Observacoes ...:
--------------------------------------------------------------------------------------------------*/
Static Function EditaTrocaSenha( nLin, nCol, cUsuario, cSenhaAtual, cNovaSenha )
   Local cConfSenha
   Local oGet
   Local aGets := {}
   Local cUsuGet

   cNovaSenha := Space( 32 )
   cConfSenha := Space( 32 )

   nLin++
   /* Se o usuario foi informado, exibe na tela */
   If ! Empty( cUsuario )
      cUsuGet := PadR( cUsuario, 50 )
      oGet := GetNew( nLin++, nCol + 23, { | x | IIf( x == NIL, cUsuGet, cUsuGet := x ) }, "cUsuGet", "@K!S30" )
      oGet:PreBlock := { || .F. }
      AAdd( aGets, oGet )
   EndIf

   // Solicita a nova senha
   oGet := GetNew( nLin++, nCol + 23, { | x | IIf( x == NIL, cNovaSenha, cNovaSenha := x ) }, "cNovaSenha", "@S30" )
   oGet:PreBlock := { || DispOutAt( 24, 1, "Informe sua nova senha!                         ", .T. ) }
   oGet:PostBlock := { || ! Empty( cNovaSenha ) }
   AAdd( aGets, oGet )

   // Solicita a confirmacao da nova senha
   oGet := GetNew( nLin, nCol + 23, { | x | IIf( x == NIL, cConfSenha, cConfSenha := x ) }, "cConfSenha", "@S30" )
   oGet:PreBlock := { || DispOutAt( 24, 1, "Confirme sua nova senha!                         ", .T. ) }
   oGet:PostBlock := { || ! Empty( cConfSenha ) }
   AAdd( aGets, oGet )

   ReadModal( aGets )

   cNovaSenha := AllTrim( cNovaSenha )

Return Nil

Function __DBGENTRY()
Return Nil

#PRAGMA BEGINDUMP


#include "cfl_lock.h"
#include <stdlib.h>

#ifdef __linux__

   #include <stdio.h>
   #include <unistd.h>
   #include <errno.h>
   #include <regex.h>
   #include <dirent.h>
   #include <fcntl.h>
   #include <linux/input.h>
   #include <stdbool.h>
   #include <stdint.h>
   #include <sys/time.h>

#else
   #include "windows.h"
#endif

   #include "cfl_thread.h"
   #include "hbapi.h"


#ifdef __XHB__
   #define HB_SIZE ULONG
   #define HB_BOOL BOOL
#endif

static FILE *s_logHandle = NULL;
static CFL_LOCKP s_lock = NULL;

static int s_loopTeclas = 1;
static char *s_buffer;
static int s_bufferLen = 0;
static int s_freeBuffer = 0;
static unsigned int s_intervalo;

HB_FUNC( F_GETPID )
{
   #if defined(_MSC_VER) || defined(__BORLANDC__)
      hb_retnint( GetCurrentProcessId() );
   #elif ( defined( HB_OS_OS2 ) && defined( __GNUC__ ) )
      hb_retnint( _getpid() );
   #else
      hb_retnint( getpid() );
   #endif
}

HB_FUNC( OPENLOG )
{
   PHB_ITEM pFileLog = hb_param( 1, HB_IT_STRING );

   if (s_logHandle == NULL && pFileLog != NULL && hb_itemGetCLen( pFileLog ) > 0) {
      s_lock = cfl_lock_new();
      s_logHandle = fopen(hb_itemGetCPtr( pFileLog ), "ab");
   }
}

HB_FUNC( LOGINFO ) {
   int iParCount = hb_pcount();
   int i;
   PHB_ITEM pItem;
   HB_SIZE nLen;
   HB_BOOL bFreeStr;
   char *str;
   char *strOri;

   if (s_logHandle && iParCount > 0) {
      cfl_lock_acquire(s_lock);
      for (i = 0; i < iParCount; i++) {
         pItem = hb_param(i + 1, HB_IT_ANY);
         if (pItem) {
            if (HB_IS_STRING(pItem)) {
               fprintf(s_logHandle,"%.*s", (int) hb_itemGetCLen(pItem), hb_itemGetCPtr(pItem));
            } else {
               strOri = str = hb_itemString(pItem, &nLen, &bFreeStr);
               while (nLen > 0 && *str == ' ') {
                  ++str;
                  --nLen;
               }
               fprintf(s_logHandle,"%.*s", (int) nLen, str);
               if (bFreeStr) {
                  hb_xfree(strOri);
               }
            }
         }
      }
      fprintf(s_logHandle,"\n");
      fflush(s_logHandle);
      cfl_lock_release(s_lock);
   }
}

HB_FUNC( CLOSELOG )
{
   if (s_lock) {
      cfl_lock_free(s_lock);
   }
   if (s_logHandle) {
       fclose(s_logHandle);
   }
}

#ifndef __HBR__
   HB_FUNC(HB_OSPATHSEPARATOR)
   {
      HB_FUNC_EXEC(HB_PS);
   }

   HB_FUNC(GETENV)
   {
      HB_FUNC_EXEC(HB_GETENV);
   }
#endif


#ifdef __linux__
   static void simulateKeyPress(int conIn, int key) {
     struct input_event forcedKey;
     forcedKey.type = EV_KEY;
     forcedKey.value = 1;    // Press
     forcedKey.code = key;
     gettimeofday(&forcedKey.time,NULL);
     write(conIn, &forcedKey, sizeof(struct input_event));
   }
#else
   static void simulateKeyPress(HANDLE conIn, int key) {
      DWORD dw;
      INPUT_RECORD record[2];
      record[0].EventType = KEY_EVENT;
      record[0].Event.KeyEvent.bKeyDown = TRUE;
      record[0].Event.KeyEvent.dwControlKeyState = 0;
      record[0].Event.KeyEvent.uChar.AsciiChar = key;
      record[0].Event.KeyEvent.wRepeatCount = 1;
      record[0].Event.KeyEvent.wVirtualKeyCode = toupper(key); /* virtual keycode is always uppercase */
      record[0].Event.KeyEvent.wVirtualScanCode = MapVirtualKeyA(key & 0x00ff, 0);
      WriteConsoleInput(conIn, record, 1 , &dw);
   }
#endif

   static void loopTeclas(void *param) {
#ifdef __linux__
      int conIn = STDIN_FILENO;
//      int conIn = -1;
//      char *dirName = "/dev/input/by-id";
//      DIR *dirp;
//      struct dirent *dp;
//      char fullPath[1024];
//      int result;
//      regex_t kbd;
//
//      if(regcomp(&kbd,"event-kbd",0)!=0) {
//         printf("regcomp for kbd failed\n");
//         return 0;
//      }
//
//      if ((dirp = opendir(dirName)) == NULL) {
//         perror("couldn't open '/dev/input/by-path'");
//         return 0;
//      }
//
//
//      do {
//         errno = 0;
//         if ((dp = readdir(dirp)) != NULL) {
//            if(regexec (&kbd, dp->d_name, 0, NULL, 0) == 0) {
//               sprintf(fullPath,"%s/%s",dirName,dp->d_name);
//               conIn = open(fullPath,O_WRONLY | O_NONBLOCK);
//               result = ioctl(conIn, EVIOCGRAB, 1);
//            }
//
//         }
//      } while (dp != NULL);
//
//      closedir(dirp);
//
//
//      regfree(&kbd);

#else
      HANDLE conIn = GetStdHandle(STD_INPUT_HANDLE);
#endif
      int i = 0;
      while (s_loopTeclas) {
         simulateKeyPress(conIn, (int) s_buffer[ i % s_bufferLen ]);
         ++i;
         #ifdef __linux__
            sleep( s_intervalo );
         #else
            Sleep( s_intervalo );
         #endif
      }
      if (s_freeBuffer) {
         hb_xfree(s_buffer);
      }
#ifdef __linux__
      close(conIn);
#endif
   }

   HB_FUNC( SIMULA_INTERACAO )
   {
      CFL_THREADP hThread;
      PHB_ITEM pBuffer = hb_param( 1, HB_IT_STRING );
      PHB_ITEM pIntervalo = hb_param( 2, HB_IT_NUMERIC );

      if (pBuffer == NULL || hb_itemGetCLen(pBuffer) == 0) {
         s_buffer = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
         s_bufferLen = 26;
         s_freeBuffer = 0;
      } else {
         s_bufferLen = (int) hb_itemGetCLen(pBuffer);
         s_buffer = ( char * ) hb_xgrab( s_bufferLen + 1 );
         strncpy(s_buffer, hb_itemGetCPtr(pBuffer), s_bufferLen);
         s_buffer[ s_bufferLen ] = '\0';
         s_freeBuffer = 1;
      }

      if (pIntervalo == NULL || hb_itemGetNI(pIntervalo) == 0) {
         s_intervalo = 100;
      } else {
         s_intervalo = hb_itemGetNI(pIntervalo);
      }

      s_loopTeclas = 1;
      hThread = cfl_thread_new(loopTeclas);
      cfl_thread_start(hThread, NULL);
   }

   HB_FUNC( INTERROMPE_SIMULACAO_INTERACAO )
   {
      s_loopTeclas = 0;
   }

#PRAGMA ENDDUMP
