import time
import requests
import streamlit as st

API_URL = "http://localhost:8001/api/v1/chat/"

SERVICES = {
    "mock_service_2 (CSV)":   "http://localhost:8002/api/v1/messages/",
    "mock_service_3 (Excel)": "http://localhost:8003/api/v1/messages/",
    "mock_service_4 (PDF)":   "http://localhost:8004/api/v1/messages/",
}

st.title("Agentic AI Orchestrator")

# ── Send message ──────────────────────────────────────────────────────────────
st.subheader("Send Message")
file_type = st.selectbox("File Type", ["CSV", "Excel", "PDF"])
message = st.text_input("Message")

if st.button("Send") and message.strip():
    payload = {"message": f"[{file_type}] {message}"}
    try:
        response = requests.post(API_URL, json=payload)
        response.raise_for_status()
        st.success(response.json()["reply"])
    except requests.exceptions.ConnectionError:
        st.error("Could not connect to mock_service_1 on port 8001.")
    except Exception as e:
        st.error(f"Error: {e}")

st.divider()

# ── Live service inbox ────────────────────────────────────────────────────────
st.subheader("Service Inbox")

cols = st.columns(3)

for col, (label, url) in zip(cols, SERVICES.items()):
    with col:
        st.markdown(f"**{label}**")
        try:
            resp = requests.get(url, timeout=2)
            resp.raise_for_status()
            msgs = resp.json().get("messages", [])
            if msgs:
                for m in reversed(msgs):
                    st.success(m.get("message", str(m)))
            else:
                st.info("No messages yet")
        except requests.exceptions.ConnectionError:
            st.warning("Service offline")
        except Exception as e:
            st.error(str(e))

time.sleep(3)
st.rerun()
