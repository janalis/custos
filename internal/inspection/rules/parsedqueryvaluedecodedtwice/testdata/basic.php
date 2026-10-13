<?php
parse_str("next=a%252Fb",$q);$v=<warning descr="Use the query value without decoding it a second time.">urldecode($q["next"])</warning>;
