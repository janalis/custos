<?php
if(openssl_verify($s,$sig,$key,OPENSSL_ALGO_SHA256)===1){echo "accepted";}
