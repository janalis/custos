<?php
$re1 = '~' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($needle) . '~i';
$re2 = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error> ( $host );
$re3 = '%' . \<error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>(...$parts) . '%';
