# Deploying to Microsoft Azure (free account) with beyondexpeditions.in

Same result as the Oracle guide ([DEPLOY.md](DEPLOY.md)): website with your domain and automatic
HTTPS, admin console kept private. **Only the "create the server" and "open the ports" steps are
different.** Steps 5 to 11 of DEPLOY.md (DNS, deploy, admin sign-in, before-launch checks, updates,
backups) are the same; the differences are listed at the end.

> Azure's portal changes often. If a menu name isn't an exact match, look for the closest one.

## Know these three things first

1. **The free VM lasts 12 months.** After that the server is billed at the normal price (a few dollars a month). Put a reminder in your calendar for month 11 to either pay, or move the site.
2. **The $200 credit lasts only 30 days, and then you must upgrade to "Pay-As-You-Go".** The account asks you to do this. It is not a trap: the free VM hours stay free for the 12 months after upgrading. But it means Azure can now bill your card for anything you create beyond the free allowance, so create a **budget alert** (step 5).
3. **The free VM sizes are only free with the right disk.** Microsoft's current wording: 750 hours a month of **B1s**, **B2ats v2** (AMD) or **B2pts v2** (Arm), and the VM is free only with a **64 GiB (P6) managed disk**. B1s is often "not available" for free accounts, so this guide uses **B2ats v2**.

## 1. Make the SSH key

You already have one on this PC: `C:\Users\<you>\.ssh\oracle.key` (the name doesn't matter). Its public half is in `oracle.key.pub`; you paste that into Azure in step 3.

## 2. Check what you can create

Portal > **Subscriptions** > your subscription > **Usage + quotas**. Search "B2ats" (or "Bsv2") and pick a region near you: **Central India (Pune)**, **South India (Chennai)** or **West India (Mumbai)**. A free account sometimes has zero quota in a region or says a size is out of capacity. If so, try the next region.

## 3. Create the virtual machine

Portal > **Virtual machines > Create > Azure virtual machine**.

| Setting | Choose |
|---|---|
| Region | Central India, South India or West India (the one that works) |
| Availability options | No infrastructure redundancy required |
| Image | **Ubuntu Server 24.04 LTS** (x64 Gen2). For the Arm size B2pts v2 pick the Arm64 image instead |
| Size | **Standard_B2ats_v2** (AMD, 1 GiB RAM). It shows a "free services eligible" style label if it qualifies |
| Authentication type | **SSH public key** |
| Username | `azureuser` |
| SSH public key source | *Use existing public key*, and paste the line from `oracle.key.pub` |
| Public inbound ports | *Allow selected ports*: **SSH (22), HTTP (80), HTTPS (443)**. Do not add 8080 or 8081 |

**Disks tab:** OS disk type **Premium SSD (locally-redundant)**, size **64 GiB** (this is the "P6" the free offer covers). Leave *Delete with VM* ticked.

**Networking tab:** keep the new virtual network. For **Public IP**, create new and make sure it is **Static** (a Standard SKU address is always static), so the address never changes.

**Management tab:** turn **off** auto-shutdown (otherwise the site goes offline every evening).

On **Review + create**, look at the price. It should show about **$0.00** for the VM and disk under your free account. If it shows a real monthly price, go back and fix the size or disk before creating. Then **Create**.

When it finishes, open the VM and note its **public IP address**.

Test the login (answer `yes` to the fingerprint question):
```powershell
ssh -i $HOME\.ssh\oracle.key azureuser@<the-ip>
```

## 4. Check the ports

In the VM page: **Networking** (or *Network settings*). Under inbound rules you should see 22, 80 and 443 allowed. If 80 or 443 is missing, add an inbound rule: any source, destination port 80 (and again for 443), protocol TCP, action Allow.

## 5. Put a spending limit alert on the account

Portal > **Cost Management + Billing > Cost Management > Budgets > Add**. Set the amount to **$5** with an email alert at 50% and 100%. This is how you notice any accidental charge within hours.

## Then follow DEPLOY.md

- **Step 5 (DNS):** add the two A records (`@` and `www`) pointing at the Azure public IP.
- **Step 6 (deploy):** use `azureuser` instead of `ubuntu`:
  ```powershell
  .\deploy\deploy.ps1 -Server azureuser@<the-ip> -Key $HOME\.ssh\oracle.key -Domain beyondexpeditions.in -Setup -Photos
  ```
  If you chose the Arm size (B2pts v2), add `-Arch arm64`.
- **Steps 7 to 10** are identical (replace `ubuntu@` with `azureuser@` in every command).
- **Skip DEPLOY.md step 11** (Oracle's idle-server rule). The Azure equivalent is the budget alert above.

## Troubleshooting

| Problem | What to check |
|---|---|
| "SKU not available" / "capacity" when creating | Try another India region, or another free size (B2pts v2). |
| Estimated cost isn't $0 | Wrong disk (must be the 64 GiB Premium SSD) or a size that isn't in the free list. |
| Browser can't connect | Inbound rules for 80/443 (step 4) and that DNS points at the Azure IP. |
| Account says it is disabled after 30 days | Upgrade to Pay-As-You-Go when prompted (item 2 above). |

## Reminders and cost alerts (set these up once)

**Calendar:** double-click `deploy/azure-reminders.ics` to add four reminders to Outlook, Google Calendar or Apple Calendar:
- 20 Nov 2026: the $200 credit ends, Azure asks you to upgrade to Pay-As-You-Go
- 22 Aug 2027: **10 months in, plan the move** to a cheap VPS
- 22 Sep 2027: move the site this week
- 15 Oct 2027: the free year ends in a week

They assume the account was opened on 22 Oct 2026. If the Azure portal (Subscriptions) shows a different date, shift them to match.

**An email for every $5 of cost:** Portal > **Cost Management + Billing > Cost Management > Budgets > Add**.
- Scope: your subscription. Name: `beyond-every-5-dollars`
- Reset period: **Annually** (so it adds up across the whole year). Amount: **$25**.
- Alert conditions (type *Actual*, percentage of budget) and your email as the recipient:
  **20%** ($5), **40%** ($10), **60%** ($15), **80%** ($20), **100%** ($25).
- Optional second budget: Monthly, **$5**, alert at 100%, to catch any single month that costs more than $5.

Budget alerts only **notify** you; they do not stop spending, and Azure's cost data can lag by many hours. If you ever get one, open Cost Management > Cost analysis to see what is costing money (usually the public IP, or a disk or VM left over from a test).
