create table devices (
  addr text primary key,
  name text,
  mac text,
  discoveredat timestamp,
  discoveredby text,
  -- Meta
  metadnsname text,
  metamanufacturer text,
  metatags text,
  -- Server
  serverports text,
  serverlastscan timestamp,
  -- PerfPing
  perfpingfirstseen timestamp,
  perfpinglastseen timestamp,
  perfpingmeanping integer,
  perfpingmaxping integer,
  perfpinglastfailed integer,
  -- Snmp
  snmpname text,
  snmpdescription text,
  snmpcommunity text,
  snmpport integer,
  snmplastcheck timestamp,
  snmphasarptable text,
  snmplastarptablescan timestamp,
  snmphasinterfaces text,
  snmplastinterfacesscan timestamp
);

create table networks (
  prefix text primary key,
  name string,
  lastscan timestamp,
  tags text
);

create table flows (
  start timestamp,
  end timestamp,
  srcaddr text,
  srcport integer,
  srcasn text,
  dstaddr text,
  dstport integer,
  dstasn text,
  protocol text,
  bytes integer,
  packets integer
);

create table performancepings (
  start timestamp,
  addr text,
  minimum integer,
  average integer,
  maximum integer,
  loss float
);

create table asns (
  asn text primary key,
  country text,
  name text,
  iprange text,
  created timestamp
)
