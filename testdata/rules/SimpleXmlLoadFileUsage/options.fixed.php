<?php
// F2: the fix is withheld when the options may need the document's base URI.
$r0 = simplexml_load_file($p, 'SimpleXMLElement', LIBXML_NOENT);
$r1 = simplexml_load_file($p, null, LIBXML_NOCDATA | LIBXML_DTDLOAD);
$r2 = simplexml_load_file($p, null, LIBXML_DTDATTR);
$r3 = simplexml_load_file($p, null, LIBXML_DTDVALID | LIBXML_NOERROR);
$r4 = simplexml_load_file($p, null, LIBXML_XINCLUDE);
$r5 = simplexml_load_file($p, null, $flags);
$r6 = simplexml_load_file($p, null, flags());
$r7 = simplexml_load_file($p, null, Config::FLAGS);
$r8 = simplexml_load_file($p, null, MY_FLAGS);
$r9 = simplexml_load_file($p, null, 2);
$r10 = simplexml_load_file($p, null, 0x4000 | 0x10);
$r11 = simplexml_load_file($p, null, 1.0);
$r12 = simplexml_load_file($p, null, '0');
$r13 = simplexml_load_file($p, null, LIBXML_NOCDATA + LIBXML_NOBLANKS);
$r14 = simplexml_load_file($p, null, ~LIBXML_NOENT);
$r15 = simplexml_load_file($p, null, \Lib\LIBXML_NOCDATA);
$r16 = simplexml_load_string(file_get_contents($p), null, 0);
$r17 = simplexml_load_string(file_get_contents($p), null, 0x4000);
$r18 = simplexml_load_string(file_get_contents($p), null, (LIBXML_NOCDATA | LIBXML_NOBLANKS) | 32);
$r19 = simplexml_load_string(file_get_contents($p), null, \LIBXML_NONET, 'ns', true);
