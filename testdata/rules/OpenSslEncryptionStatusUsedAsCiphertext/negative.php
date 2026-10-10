<?php
if (openssl_public_encrypt($data, $out, $key)) { echo base64_encode($out); }
