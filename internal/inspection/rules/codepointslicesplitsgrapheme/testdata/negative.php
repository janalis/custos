<?php
$s=grapheme_substr("e\u{0301}",0,1);
$zwsp=mb_substr("\u{200B}\u{301}x",0,1,"UTF-8");
$line=mb_substr("\u{2028}\u{301}x",0,1,"UTF-8");
$shy=mb_substr("\u{AD}\u{301}x",0,1,"UTF-8");
