# Crawler para coleta dos atos normativos da Agência Nacional do Petróleo, Gás Natural e Biocombustíveis

**site**: [https://atosoficiais.com.br/anp](https://atosoficiais.com.br/anp)

A coleta pode ser parametrizada por data de publicação e por tipo de ato. 

### 1. Tipo de ato

- Ata (`-flag`)
- Autorização (`-autorização`)
- Despacho (`-despacho`)
- IN Financeira Administrativa (`-in_fin_adm`)
- IN Gestão Interna (`-in_ges_interna`)
- IN Gestão Técnica (`-in_ges_tec`)
- IN Recursos Humanos (`-in_rec_humanos`)
- IN Segurança Operacional (`-in_seg_op`)
- Instrução Normativa (`-in_instr_norm`)
- Portaria ANP (`-port_anp`)
- Portaria Conjunta (`-port_conj`)
- Portaria Técnica (`-port_tecnica`)
- Resolução (`-resolução`)
- Resolução Conjunta (`-res_conjunta`)
- Resolução de Diretoria RD (`-res_diretoria`)
- Todas os atos (`-all`)

### 2. Data
As flags para determinar o intervalo de datas a serem pesquisas devem ser especificadas pelo parâmetro e pelo respectiva data no formato `%d-%m-%Y`
#### 2.1 -data_inicio
Exemplo: `-data_inicio 01-01-2020`
#### 2.2 -data_fim
Exemplo: `-data_fim 31-12-2020`

### 3. Exemplos
#### 3.1 Coletar todos os atos publicados no ano de 2020
`go run atos-anp.go -all -data_inicio 01-01-2020 -data_fim 31-12-2020`

#### 3.2 Coletar todas as Resoluções de Diretoria RD publicadas em 2019
`go run atos-anp.go -res_diretoria -data_inicio 01-01-2020 -data_fim 31-12-2020`

#### 3.3 Coletar todas as Resoluções, Autorizações e Portarias ANP publicadas entre 2015 e 2020.
`go run atos-anp.go -resolução -autorização -port_anp -data_inicio 01-01-2015 -data_fim 31-12-2020`
