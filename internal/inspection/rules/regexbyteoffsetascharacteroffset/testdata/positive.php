<?php
preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE); echo <warning descr="Convert byte offsets before using character-based slicing.">mb_substr('ñx',$m[0][1],1)</warning>;
