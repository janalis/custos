<?php
$s=<warning descr="Slice display text at grapheme boundaries.">mb_substr("e\u{0301}",0,1,"UTF-8")</warning>;
$zwj=<warning descr="Slice display text at grapheme boundaries.">mb_substr("\u{200D}\u{301}x",0,1,"UTF-8")</warning>;
$tag=<warning descr="Slice display text at grapheme boundaries.">mb_substr("a\u{E0061}\u{301}x",0,2,"UTF-8")</warning>;
