<?php
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (!<warning descr="Compare cURL transport failure strictly with false.">curl_exec($ch)</warning>) { throw new RuntimeException(); }
