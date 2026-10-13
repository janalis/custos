<?php
echo "one", "two"; echo "ok"; $f=new NumberFormatter("en_US",NumberFormatter::DECIMAL);$n=$f->parse($unknown);echo $n;
