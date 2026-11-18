# Threat Model and Defense Mechanics

This document outlines the threat vectors targeted by honeyagent and the security guarantees provided by its deception architecture.

## Target threat vectors

### 1. Indirect prompt injection and knowledge exfiltration
- **Attack mechanism**: An adversary injects jailbreak prompts into an agent or asks targeted questions to trigger retrieval of sensitive internal documents, such as executive payroll, cloud credentials, or unannounced corporate transactions.
- **Deception countermeasure**: Synthetic lure documents seeded across vector indices contain attractive keywords (for example: `AWS_SECRET_ACCESS_KEY`, `executive bonus`, `acquisition valuation`). When the RAG pipeline retrieves a lure chunk, the stream scanner detects the embedded HMAC token, suppresses the leak, trips the alarm, and reroutes the attacker to the tarpit.

### 2. Autonomous agent tool probing and privilege escalation
- **Attack mechanism**: Attackers manipulate an agent into discovering and executing unadvertised administrative tools, attempting to run raw SQL queries, export unmasked PII, or escalate cloud permissions.
- **Deception countermeasure**: honeyagent registers administrative decoy tools in the agent tool manifest (`admin_execute_raw_sql`, `export_customer_pii`, `aws_iam_elevate_privilege`). Because legitimate workflows have no knowledge of these functions, any call is a deterministic attack signal. The interceptor logs the attempt, notifies SecOps, and returns simulated output to maintain containment.

### 3. Automated reconnaissance and brute-force scraping
- **Attack mechanism**: Attackers script rapid prompt queries to map the boundaries of internal knowledge bases.
- **Deception countermeasure**: Quarantined sessions enter the tarpit, which applies randomized latency delays between 800ms and 2500ms while returning synthetic hallucinations. This slows down automated tools and drains compute resources without tipping off the adversary.

## False positive elimination

Traditional security products rely on heuristic regexes or semantic classifiers that produce false alarms on benign queries containing words like "database" or "salary".

honeyagent achieves zero false alarms through cryptographic separation:
- Legitimate enterprise documents do not contain valid HMAC canary signatures signed by the master key.
- Normal business applications never invoke unadvertised administrative decoy tools.
- Detection triggers only on a verified cryptographic signature match or a direct decoy function call.
