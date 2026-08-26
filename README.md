# SheepContabil — Portal de Automações

Portal demonstrativo do desafio técnico da Sheep Technology. A entrega implementa quatro processos completos em uma aplicação única:

- **SC-02:** painel de situação fiscal dos clientes (RPA de alta complexidade);
- **SC-05:** bloqueio e desbloqueio de clientes inadimplentes (RPA);
- **SC-06:** briefing societário com perguntas condicionais (controle sistematizado);
- **SC-20:** vencimento de certificados digitais (controle sistematizado).

## Acesso de demonstração

| Perfil | E-mail | Senha | Acesso |
|---|---|---|---|
| Administrador | `admin@sheepcontabil.com` | `Sheep@2026` | Todos os módulos |
| Operador | `operador@sheepcontabil.com` | `Sheep@2026` | SC-06 e SC-20 |

Todos os nomes, documentos e eventos são sintéticos.

## Executar localmente

Requisitos: Go 1.24+, Node 18+ e npm.

```bash
cd web
npm install
npm run build
cd ..
go run ./cmd/server
```

Acesse `http://localhost:8080`. A base inicial é criada automaticamente em `data/sheep.json`.

Também é possível executar com:

```bash
docker compose up --build
```

## Decisões e suposições

### Arquitetura

O backend é um monólito modular em Go e o frontend é React. Essa escolha reduz infraestrutura e pontos de falha sem misturar as fronteiras: cada integração externa do SC-05 aparece como um adaptador simulado independente. A persistência local em JSON foi escolhida para tornar a demonstração autocontida; em produção, o `Store` seria substituído por PostgreSQL sem alterar os handlers ou as regras.

O backend está organizado por responsabilidade em `internal/app`: infraestrutura HTTP, autenticação, dashboard, agendador, persistência e um arquivo para cada processo. No frontend, `pages` contém as telas, `components` reúne elementos reutilizáveis, `hooks` concentra comportamento compartilhado e `lib` isola o cliente HTTP.

### SC-05

- Assumimos três sistemas sem API comum: gestão contábil, financeiro e tarefas.
- A automação executa a sequência ordenada e interrompe diante de falha.
- No sistema de tarefas, o cliente não é excluído: o responsável anterior é preservado e trocado por `BLOQUEADO`, exatamente como solicitado.
- O checkbox de falha torna demonstrável o comportamento diante da indisponibilidade de um terceiro.
- Em produção, cada item da sequência chamaria um adaptador real com credencial em secret manager.

### SC-02

- Assumimos consultas mensais à Receita Federal, FGTS e Secretaria Estadual.
- Cada combinação de cliente e órgão gera um registro com horário, resultado e número de tentativas.
- “Consulta falhou” é diferente de “regular”: a falha permanece visível e não produz conclusão fiscal falsa.
- A execução pode ser manual ou mensal automática; uma opção demonstra três tentativas diante da indisponibilidade do FGTS.
- Em produção, cada órgão seria um adaptador isolado, permitindo substituir a automação do portal sem alterar o painel.

### SC-06

- As condicionais são registros de `BriefingRule`, não condicionais espalhadas pela interface.
- Cliente fora de Alagoas exige inscrição estadual e município.
- Sócio casado exige regime de casamento.
- Alteração contratual exige a descrição da alteração.
- O servidor repete todas as validações; ocultar um campo no navegador não permite concluir um briefing incompleto.

### SC-20

- A janela de antecedência adotada é de 60 dias.
- O portal separa certificados em até 30 dias, entre 31 e 60 e fora da janela.
- Um aviso registra data e destinatário. Assim, a operação vê o que mudou e decide conscientemente se deve reenviar.
- A varredura pode ser disparada manualmente e também roda pelo agendador interno. Na inicialização e a cada seis horas, o serviço garante uma única execução mensal idempotente.

## Segurança e falhas

- Sessões aleatórias em cookie `HttpOnly` e `SameSite=Lax`;
- autorização verificada no servidor por perfil e módulo;
- limite de 1 MB e validação dos corpos JSON;
- mensagens operacionais legíveis, sem stack trace exposto;
- gravação atômica da base por arquivo temporário;
- cabeçalhos básicos de proteção e segredos fora do repositório.

Em produção, senhas devem usar Argon2id ou bcrypt; o SHA-256 com salt fixo existe somente para não adicionar dependência ao ambiente demonstrativo.

## Testes

```bash
go test ./...
cd web && npm run build
```

## Uso de IA

Um assistente de IA foi usado para apoiar o desenho da solução, geração inicial de código e documentação. Nenhuma IA é usada durante a execução das automações e nenhum dado é enviado a serviços externos.
