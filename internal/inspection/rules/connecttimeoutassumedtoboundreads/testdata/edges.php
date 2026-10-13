<?php
fgets($unknown);$s=fopen("x","r");fgets($s);$s=fsockopen("host",80,$errno,$error,0);fgets($s);$s=fsockopen("host",80,$errno,$error,2);if(stream_set_timeout($s,1)){fgets($s);}$t=fsockopen("host",80,$errno,$error,2);if(stream_set_timeout($t,1)){fgets($t);}
