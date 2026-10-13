<?php
$x = new SimpleXMLElement('<r/>');
<warning descr="Escape bare ampersands in XML child values.">$x->addChild('a', 'q=1&r=2;')</warning>;
<warning descr="Escape bare ampersands in XML child values.">$x->addChild('b', '&1st;')</warning>;
<warning descr="Escape bare ampersands in XML child values.">$x->addChild('c', '&-x;')</warning>;
<warning descr="Escape bare ampersands in XML child values.">$x->addChild('d', '&;')</warning>;
<warning descr="Escape bare ampersands in XML child values.">$x->addChild('e', "a &\n b; c")</warning>;
<warning descr="Escape bare ampersands in XML child values.">$x->addChild('f', '&a&amp;')</warning>;
<warning descr="Escape bare ampersands in XML child values.">$x->addChild('g', '&a×b;')</warning>;
<warning descr="Escape bare ampersands in XML child values.">$x->addChild('h', '&amp; then x&y=1;')</warning>;
$x->addChild('i', '&company.name; &_a-1; &ns:e; &été; &日本; &a·b;');
$x->addChild('j', "&\xFF;");
